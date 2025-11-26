output "k3s_master_private_ip" {
  description = "Private IP of k3s master node"
  value       = aws_instance.k3s_master[0].private_ip
}

output "k3s_master_id" {
  description = "Instance ID of k3s master"
  value       = aws_instance.k3s_master[0].id
}

output "k3s_worker_private_ips" {
  description = "Private IPs of k3s worker nodes"
  value       = aws_instance.k3s_workers[*].private_ip
}

output "k3s_worker_ids" {
  description = "Instance IDs of k3s workers"
  value       = aws_instance.k3s_workers[*].id
}

output "alb_dns_name" {
  description = "DNS name of the Application Load Balancer"
  value       = aws_lb.main.dns_name
}

output "alb_arn" {
  description = "ARN of the Application Load Balancer"
  value       = aws_lb.main.arn
}

output "alb_target_group_arn" {
  description = "ARN of the ALB target group"
  value       = aws_lb_target_group.app.arn
}