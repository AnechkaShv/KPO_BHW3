import React, { useState } from 'react';

const Order = () => {
  const [userId, setUserId] = useState('test_user');
  const [amount, setAmount] = useState('');
  const [description, setDescription] = useState('');
  const [orderId, setOrderId] = useState('');
  const [orders, setOrders] = useState([]);
  const [message, setMessage] = useState('');
  const [loading, setLoading] = useState(false);

  const createOrder = async () => {
    if (!userId || !amount || !description) {
      setMessage('All fields are required');
      return;
    }
    
    setLoading(true);
    try {
      const response = await fetch('http://localhost:8000/orders/create', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          user_id: userId,
          amount: parseFloat(amount),
          description
        })
      });
      
      if (!response.ok) throw new Error('Failed to create order');
      
      const data = await response.json();
      setOrderId(data.order_id);
      setMessage('Order created successfully!');
    } catch (error) {
      setMessage(error.message);
    } finally {
      setLoading(false);
    }
  };

  const getOrder = async () => {
    if (!orderId) {
      setMessage('Order ID is required');
      return;
    }
    
    setLoading(true);
    try {
      const response = await fetch(`http://localhost:8000/orders/get?id=${orderId}`);
      
      if (!response.ok) throw new Error('Failed to fetch order');
      
      const data = await response.json();
      setOrders([data]);
      setMessage('');
    } catch (error) {
      setMessage(error.message);
    } finally {
      setLoading(false);
    }
  };

  const listOrders = async () => {
    if (!userId) {
      setMessage('User ID is required');
      return;
    }
    
    setLoading(true);
    try {
      const response = await fetch(`http://localhost:8000/orders/list?user_id=${userId}`);
      
      if (!response.ok) throw new Error('Failed to fetch orders');
      
      const data = await response.json();
      setOrders(data);
      setMessage('');
    } catch (error) {
      setMessage(error.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="order-container">
      <h2>Order Management</h2>
      
      <div className="form-group">
        <input
          type="text"
          placeholder="User ID"
          value={userId}
          onChange={(e) => setUserId(e.target.value)}
        />
      </div>
      
      <div className="form-group">
        <input
          type="number"
          placeholder="Amount"
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
        />
      </div>
      
      <div className="form-group">
        <input
          type="text"
          placeholder="Description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
        />
      </div>
      
      <div className="action-buttons">
        <button onClick={createOrder} disabled={loading}>
          {loading ? 'Processing...' : 'Create Order'}
        </button>
      </div>
      
      <div className="form-group">
        <input
          type="text"
          placeholder="Order ID"
          value={orderId}
          onChange={(e) => setOrderId(e.target.value)}
        />
        <button onClick={getOrder} disabled={loading}>
          {loading ? 'Processing...' : 'Get Order'}
        </button>
      </div>
      
      <div className="action-buttons">
        <button onClick={listOrders} disabled={loading}>
          {loading ? 'Processing...' : 'List Orders'}
        </button>
      </div>
      
      {orders.length > 0 && (
        <div className="orders-list">
          <h3>Orders</h3>
          {orders.map(order => (
            <div key={order.order_id} className="order-item">
              <p><strong>ID:</strong> {order.order_id}</p>
              <p><strong>Amount:</strong> ${order.amount}</p>
              <p><strong>Description:</strong> {order.description}</p>
              <p><strong>Status:</strong> {order.status}</p>
            </div>
          ))}
        </div>
      )}
      
      {message && <div className={`message ${message.includes('failed') ? 'error' : ''}`}>
        {message}
      </div>}
    </div>
  );
};

export default Order;