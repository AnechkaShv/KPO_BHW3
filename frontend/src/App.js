// frontend/src/App.js
import React, { useState, useEffect } from 'react';
import CreateOrderForm from './components/CreateOrderForm';
import OrderList from './components/OrderList';
import { getOrders } from './api';
import './App.css';

function App() {
  const [orders, setOrders] = useState([]);
  const [userId] = useState('test_user'); // Можно генерировать динамически

  const fetchOrders = async () => {
    try {
      const ordersData = await getOrders(userId);
      setOrders(ordersData);
    } catch (error) {
      console.error('Failed to fetch orders:', error);
    }
  };

  useEffect(() => {
    fetchOrders();
  }, [userId]);

  const handleOrderCreated = (newOrder) => {
    setOrders([...orders, newOrder]);
  };

  return (
    <div className="app">
      <h1>E-Commerce System</h1>
      <div className="container">
        <CreateOrderForm onOrderCreated={handleOrderCreated} />
        <OrderList orders={orders} />
      </div>
    </div>
  );
}

export default App;