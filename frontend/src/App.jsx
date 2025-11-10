import React, { useState, useEffect } from 'react';
import { ShoppingCart, User, Package, Bell, LogOut, Search, Plus, Minus, Trash2, CheckCircle } from 'lucide-react';

// API Configuration
const API_BASE = 'http://localhost:8080/api/v1';

// Main App Component
const App = () => {
  const [currentPage, setCurrentPage] = useState('login');
  const [user, setUser] = useState(null);
  const [token, setToken] = useState(localStorage.getItem('token'));
  const [cart, setCart] = useState([]);
  const [products, setProducts] = useState([]);
  const [orders, setOrders] = useState([]);
  const [notifications, setNotifications] = useState([]);
  const [searchTerm, setSearchTerm] = useState('');

  // Authentication
  useEffect(() => {
    if (token) {
      fetchUserProfile();
      if (currentPage === 'login' || currentPage === 'register') {
        setCurrentPage('products');
      }
    }
  }, [token]);

  const fetchUserProfile = async () => {
    try {
      const response = await fetch(`${API_BASE}/users/me`, {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      const data = await response.json();
      if (data.success) setUser(data.data);
    } catch (error) {
      console.error('Failed to fetch user:', error);
    }
  };

  const handleLogin = async (email, password) => {
    try {
      const response = await fetch(`${API_BASE}/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password })
      });
      const data = await response.json();
      if (data.success) {
        setToken(data.data.token);
        localStorage.setItem('token', data.data.token);
        setUser(data.data.user);
        setCurrentPage('products');
      } else {
        alert(data.error || 'Login failed');
      }
    } catch (error) {
      alert('Login failed: ' + error.message);
    }
  };

  const handleRegister = async (email, password, fullName) => {
    try {
      const response = await fetch(`${API_BASE}/auth/register`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password, full_name: fullName })
      });
      const data = await response.json();
      if (data.success) {
        alert('Registration successful! Please login.');
        setCurrentPage('login');
      } else {
        alert(data.error || 'Registration failed');
      }
    } catch (error) {
      alert('Registration failed: ' + error.message);
    }
  };

  const handleLogout = () => {
    setToken(null);
    setUser(null);
    setCart([]);
    localStorage.removeItem('token');
    setCurrentPage('login');
  };

  // Products
  const fetchProducts = async () => {
    try {
      const response = await fetch(`${API_BASE}/products`);
      const data = await response.json();
      if (data.success) setProducts(data.data || []);
    } catch (error) {
      console.error('Failed to fetch products:', error);
    }
  };

  useEffect(() => {
    if (currentPage === 'products') fetchProducts();
  }, [currentPage]);

  const addToCart = (product) => {
    const existing = cart.find(item => item.id === product.id);
    if (existing) {
      setCart(cart.map(item => 
        item.id === product.id ? { ...item, quantity: item.quantity + 1 } : item
      ));
    } else {
      setCart([...cart, { ...product, quantity: 1 }]);
    }
  };

  const updateCartQuantity = (productId, change) => {
    setCart(cart.map(item => {
      if (item.id === productId) {
        const newQuantity = item.quantity + change;
        return newQuantity > 0 ? { ...item, quantity: newQuantity } : null;
      }
      return item;
    }).filter(Boolean));
  };

  const removeFromCart = (productId) => {
    setCart(cart.filter(item => item.id !== productId));
  };

  const placeOrder = async () => {
    if (cart.length === 0) {
      alert('Cart is empty!');
      return;
    }

    try {
      const items = cart.map(item => ({
        product_id: item.id,
        quantity: item.quantity
      }));

      const response = await fetch(`${API_BASE}/orders`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
          'X-User-ID': user?.id || 'test-user'
        },
        body: JSON.stringify({ items })
      });

      const data = await response.json();
      if (data.success) {
        alert('Order placed successfully!');
        setCart([]);
        setCurrentPage('orders');
      } else {
        alert(data.error || 'Order failed');
      }
    } catch (error) {
      alert('Order failed: ' + error.message);
    }
  };

  // Orders
  const fetchOrders = async () => {
    try {
      const response = await fetch(`${API_BASE}/orders`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'X-User-ID': user?.id || 'test-user'
        }
      });
      const data = await response.json();
      if (data.success) setOrders(data.data || []);
    } catch (error) {
      console.error('Failed to fetch orders:', error);
    }
  };

  useEffect(() => {
    if (currentPage === 'orders' && token) fetchOrders();
  }, [currentPage, token]);

  const cartTotal = cart.reduce((sum, item) => sum + (item.price * item.quantity), 0);
  const filteredProducts = products.filter(p => 
    p.name.toLowerCase().includes(searchTerm.toLowerCase())
  );

  // Render Components
  if (!token || currentPage === 'login') {
    return <LoginPage onLogin={handleLogin} onRegister={() => setCurrentPage('register')} />;
  }

  if (currentPage === 'register') {
    return <RegisterPage onRegister={handleRegister} onLogin={() => setCurrentPage('login')} />;
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 py-4 flex items-center justify-between">
          <div className="flex items-center space-x-8">
            <h1 className="text-2xl font-bold text-blue-600">E-Commerce</h1>
            <nav className="flex space-x-4">
              <NavButton icon={<Package size={18} />} label="Products" active={currentPage === 'products'} onClick={() => setCurrentPage('products')} />
              <NavButton icon={<ShoppingCart size={18} />} label={`Cart (${cart.length})`} active={currentPage === 'cart'} onClick={() => setCurrentPage('cart')} />
              <NavButton icon={<Bell size={18} />} label="Orders" active={currentPage === 'orders'} onClick={() => setCurrentPage('orders')} />
            </nav>
          </div>
          <div className="flex items-center space-x-4">
            <span className="text-sm text-gray-600">Hello, {user?.full_name}</span>
            <button onClick={handleLogout} className="flex items-center space-x-1 text-red-600 hover:text-red-700">
              <LogOut size={18} />
              <span>Logout</span>
            </button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 py-8">
        {currentPage === 'products' && (
          <ProductsPage 
            products={filteredProducts} 
            searchTerm={searchTerm}
            setSearchTerm={setSearchTerm}
            onAddToCart={addToCart} 
          />
        )}
        {currentPage === 'cart' && (
          <CartPage 
            cart={cart} 
            total={cartTotal}
            onUpdateQuantity={updateCartQuantity}
            onRemove={removeFromCart}
            onCheckout={placeOrder}
          />
        )}
        {currentPage === 'orders' && <OrdersPage orders={orders} />}
      </main>
    </div>
  );
};

// Login Page
const LoginPage = ({ onLogin, onRegister }) => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center p-4">
      <div className="bg-white rounded-lg shadow-xl p-8 w-full max-w-md">
        <h2 className="text-3xl font-bold text-center mb-8">Welcome Back</h2>
        <div className="space-y-4">
          <input
            type="email"
            placeholder="Email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-blue-500 outline-none"
          />
          <input
            type="password"
            placeholder="Password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-blue-500 outline-none"
          />
          <button
            onClick={() => onLogin(email, password)}
            className="w-full bg-blue-600 text-white py-3 rounded-lg hover:bg-blue-700 font-medium"
          >
            Login
          </button>
          <button
            onClick={onRegister}
            className="w-full bg-gray-100 text-gray-700 py-3 rounded-lg hover:bg-gray-200 font-medium"
          >
            Create Account
          </button>
        </div>
      </div>
    </div>
  );
};

// Register Page
const RegisterPage = ({ onRegister, onLogin }) => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [fullName, setFullName] = useState('');

  return (
    <div className="min-h-screen bg-gradient-to-br from-purple-500 to-pink-600 flex items-center justify-center p-4">
      <div className="bg-white rounded-lg shadow-xl p-8 w-full max-w-md">
        <h2 className="text-3xl font-bold text-center mb-8">Create Account</h2>
        <div className="space-y-4">
          <input
            type="text"
            placeholder="Full Name"
            value={fullName}
            onChange={(e) => setFullName(e.target.value)}
            className="w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-purple-500 outline-none"
          />
          <input
            type="email"
            placeholder="Email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-purple-500 outline-none"
          />
          <input
            type="password"
            placeholder="Password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-purple-500 outline-none"
          />
          <button
            onClick={() => onRegister(email, password, fullName)}
            className="w-full bg-purple-600 text-white py-3 rounded-lg hover:bg-purple-700 font-medium"
          >
            Register
          </button>
          <button
            onClick={onLogin}
            className="w-full bg-gray-100 text-gray-700 py-3 rounded-lg hover:bg-gray-200 font-medium"
          >
            Already have an account? Login
          </button>
        </div>
      </div>
    </div>
  );
};

// Products Page
const ProductsPage = ({ products, searchTerm, setSearchTerm, onAddToCart }) => (
  <div>
    <div className="mb-6 flex items-center space-x-4">
      <div className="flex-1 relative">
        <Search className="absolute left-3 top-3 text-gray-400" size={20} />
        <input
          type="text"
          placeholder="Search products..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          className="w-full pl-10 pr-4 py-3 border rounded-lg focus:ring-2 focus:ring-blue-500 outline-none"
        />
      </div>
    </div>
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      {products.map(product => (
        <div key={product.id} className="bg-white rounded-lg shadow hover:shadow-lg transition-shadow p-6">
          <div className="h-32 bg-gradient-to-br from-blue-100 to-purple-100 rounded-lg mb-4 flex items-center justify-center">
            <Package size={48} className="text-blue-600" />
          </div>
          <h3 className="text-lg font-semibold mb-2">{product.name}</h3>
          <p className="text-gray-600 text-sm mb-4">{product.description}</p>
          <div className="flex items-center justify-between">
            <span className="text-2xl font-bold text-blue-600">${product.price}</span>
            <span className="text-sm text-gray-500">Stock: {product.stock}</span>
          </div>
          <button
            onClick={() => onAddToCart(product)}
            className="w-full mt-4 bg-blue-600 text-white py-2 rounded-lg hover:bg-blue-700 flex items-center justify-center space-x-2"
          >
            <ShoppingCart size={18} />
            <span>Add to Cart</span>
          </button>
        </div>
      ))}
    </div>
  </div>
);

// Cart Page
const CartPage = ({ cart, total, onUpdateQuantity, onRemove, onCheckout }) => (
  <div className="max-w-4xl mx-auto">
    <h2 className="text-3xl font-bold mb-8">Shopping Cart</h2>
    {cart.length === 0 ? (
      <div className="text-center py-16">
        <ShoppingCart size={64} className="mx-auto text-gray-300 mb-4" />
        <p className="text-gray-500">Your cart is empty</p>
      </div>
    ) : (
      <div className="space-y-4">
        {cart.map(item => (
          <div key={item.id} className="bg-white rounded-lg shadow p-6 flex items-center space-x-6">
            <div className="w-20 h-20 bg-gray-100 rounded-lg flex items-center justify-center">
              <Package size={32} className="text-gray-400" />
            </div>
            <div className="flex-1">
              <h3 className="font-semibold text-lg">{item.name}</h3>
              <p className="text-gray-600">${item.price}</p>
            </div>
            <div className="flex items-center space-x-2">
              <button onClick={() => onUpdateQuantity(item.id, -1)} className="p-2 hover:bg-gray-100 rounded">
                <Minus size={16} />
              </button>
              <span className="w-12 text-center font-medium">{item.quantity}</span>
              <button onClick={() => onUpdateQuantity(item.id, 1)} className="p-2 hover:bg-gray-100 rounded">
                <Plus size={16} />
              </button>
            </div>
            <div className="text-lg font-semibold">${(item.price * item.quantity).toFixed(2)}</div>
            <button onClick={() => onRemove(item.id)} className="text-red-600 hover:text-red-700">
              <Trash2 size={20} />
            </button>
          </div>
        ))}
        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex justify-between items-center mb-4">
            <span className="text-xl font-semibold">Total:</span>
            <span className="text-3xl font-bold text-blue-600">${total.toFixed(2)}</span>
          </div>
          <button
            onClick={onCheckout}
            className="w-full bg-green-600 text-white py-3 rounded-lg hover:bg-green-700 font-medium flex items-center justify-center space-x-2"
          >
            <CheckCircle size={20} />
            <span>Place Order</span>
          </button>
        </div>
      </div>
    )}
  </div>
);

// Orders Page
const OrdersPage = ({ orders }) => (
  <div className="max-w-4xl mx-auto">
    <h2 className="text-3xl font-bold mb-8">Your Orders</h2>
    {orders.length === 0 ? (
      <div className="text-center py-16">
        <Package size={64} className="mx-auto text-gray-300 mb-4" />
        <p className="text-gray-500">No orders yet</p>
      </div>
    ) : (
      <div className="space-y-4">
        {orders.map(order => (
          <div key={order.id} className="bg-white rounded-lg shadow p-6">
            <div className="flex justify-between items-start mb-4">
              <div>
                <h3 className="font-semibold text-lg">Order #{order.id.substring(0, 8)}</h3>
                <p className="text-sm text-gray-500">{new Date(order.created_at).toLocaleDateString()}</p>
              </div>
              <span className={`px-3 py-1 rounded-full text-sm font-medium ${
                order.status === 'confirmed' ? 'bg-green-100 text-green-800' :
                order.status === 'pending' ? 'bg-yellow-100 text-yellow-800' :
                'bg-gray-100 text-gray-800'
              }`}>
                {order.status}
              </span>
            </div>
            <div className="border-t pt-4">
              <p className="text-2xl font-bold text-blue-600">${order.total_price.toFixed(2)}</p>
              <p className="text-sm text-gray-600 mt-2">{order.items?.length || 0} items</p>
            </div>
          </div>
        ))}
      </div>
    )}
  </div>
);

// Nav Button Component
const NavButton = ({ icon, label, active, onClick }) => (
  <button
    onClick={onClick}
    className={`flex items-center space-x-2 px-4 py-2 rounded-lg transition-colors ${
      active ? 'bg-blue-100 text-blue-700' : 'text-gray-600 hover:bg-gray-100'
    }`}
  >
    {icon}
    <span className="font-medium">{label}</span>
  </button>
);

export default App;
