# Compute Module - EC2 Instances for k3s

# Get latest Amazon Linux 2023 AMI
data "aws_ami" "amazon_linux" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["al2023-ami-*-x86_64"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

# SSH Key Pair
resource "aws_key_pair" "main" {
  key_name   = "${var.project_name}-${var.environment}-key"
  public_key = var.ssh_public_key

  tags = {
    Name        = "${var.project_name}-${var.environment}-key"
    Environment = var.environment
  }
}

# IAM Role for EC2
resource "aws_iam_role" "ec2_role" {
  name = "${var.project_name}-${var.environment}-ec2-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name        = "${var.project_name}-${var.environment}-ec2-role"
    Environment = var.environment
  }
}

# Attach policies to EC2 role
resource "aws_iam_role_policy_attachment" "ecr_read" {
  role       = aws_iam_role.ec2_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
}

resource "aws_iam_role_policy_attachment" "ssm" {
  role       = aws_iam_role.ec2_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

# Attach S3 access policy
resource "aws_iam_role_policy_attachment" "s3_access" {
  role       = aws_iam_role.ec2_role.name
  policy_arn = var.s3_access_policy_arn
}

# Instance Profile
resource "aws_iam_instance_profile" "main" {
  name = "${var.project_name}-${var.environment}-instance-profile"
  role = aws_iam_role.ec2_role.name
}

# User data script - Prepare instances for k3s
locals {
  user_data = <<-EOF
    #!/bin/bash
    set -e
    
    # Update system
    yum update -y
    
    # Install required packages for k3s
    yum install -y curl wget git
    
    # Disable firewalld (k3s will manage iptables)
    systemctl stop firewalld
    systemctl disable firewalld
    
    # Enable IP forwarding (required for k3s)
    cat <<EOT >> /etc/sysctl.conf
net.ipv4.ip_forward = 1
net.bridge.bridge-nf-call-iptables = 1
net.bridge.bridge-nf-call-ip6tables = 1
EOT
    sysctl -p
    
    # Create directory for k3s
    mkdir -p /etc/rancher/k3s
    
    # Log completion
    echo "Instance prepared for k3s at $(date)" >> /var/log/user-data.log
    
    # k3s will be installed by Ansible
  EOF
}

# K3s Master Node (first instance)
resource "aws_instance" "k3s_master" {
  count                  = 1
  ami                    = data.aws_ami.amazon_linux.id
  instance_type          = var.instance_type
  key_name               = aws_key_pair.main.key_name
  subnet_id              = var.private_subnet_ids[0]
  vpc_security_group_ids = [var.app_security_group_id]
  iam_instance_profile   = aws_iam_instance_profile.main.name

  user_data = local.user_data

  root_block_device {
    volume_type           = "gp3"
    volume_size           = var.root_volume_size
    delete_on_termination = true
    encrypted             = true
  }

  tags = {
    Name        = "${var.project_name}-k3s-master"
    Environment = var.environment
    Role        = "k3s-master"
    K3sRole     = "master" # Used by Ansible
  }
}

# K3s Worker Nodes
resource "aws_instance" "k3s_workers" {
  count                  = var.instance_count - 1 # Subtract 1 for master
  ami                    = data.aws_ami.amazon_linux.id
  instance_type          = var.instance_type
  key_name               = aws_key_pair.main.key_name
  subnet_id              = var.private_subnet_ids[count.index % length(var.private_subnet_ids)]
  vpc_security_group_ids = [var.app_security_group_id]
  iam_instance_profile   = aws_iam_instance_profile.main.name

  user_data = local.user_data

  root_block_device {
    volume_type           = "gp3"
    volume_size           = var.root_volume_size
    delete_on_termination = true
    encrypted             = true
  }

  tags = {
    Name        = "${var.project_name}-k3s-worker-${count.index + 1}"
    Environment = var.environment
    Role        = "k3s-worker"
    K3sRole     = "worker" # Used by Ansible
  }
}

# Application Load Balancer
resource "aws_lb" "main" {
  name               = "${var.project_name}-${var.environment}-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [var.alb_security_group_id]
  subnets            = var.public_subnet_ids

  enable_deletion_protection = false

  tags = {
    Name        = "${var.project_name}-${var.environment}-alb"
    Environment = var.environment
  }
}

# ALB Target Group (for k3s ingress controller)
resource "aws_lb_target_group" "app" {
  name     = "${var.project_name}-${var.environment}-tg"
  port     = 80
  protocol = "HTTP"
  vpc_id   = var.vpc_id

  health_check {
    enabled             = true
    healthy_threshold   = 2
    interval            = 30
    matcher             = "200,404" # 404 is OK if no ingress is configured yet
    path                = "/"
    port                = "traffic-port"
    protocol            = "HTTP"
    timeout             = 5
    unhealthy_threshold = 3
  }

  tags = {
    Name        = "${var.project_name}-${var.environment}-tg"
    Environment = var.environment
  }
}

# Register master with target group (k3s will run ingress controller here)
resource "aws_lb_target_group_attachment" "master" {
  target_group_arn = aws_lb_target_group.app.arn
  target_id        = aws_instance.k3s_master[0].id
  port             = 80
}

# Register workers with target group
resource "aws_lb_target_group_attachment" "workers" {
  count            = var.instance_count - 1
  target_group_arn = aws_lb_target_group.app.arn
  target_id        = aws_instance.k3s_workers[count.index].id
  port             = 80
}

# ALB Listener (HTTP)
resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.main.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.app.arn
  }
}