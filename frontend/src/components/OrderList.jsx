import React, { useEffect, useState } from 'react';
import { getOrders } from '../api';

const OrderList = ({ userId }) => {
  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchOrders = async () => {
      try {
        const data = await getOrders(userId);
        setOrders(data);
      } catch (err) {
        setError(err.message);
        console.error('Order fetch error:', err);
      } finally {
        setLoading(false);
      }
    };

    fetchOrders();
  }, [userId]);

  if (loading) return <div>Loading orders...</div>;
  if (error) return <div className="error">Error: {error}</div>;

  return (
    <div className="order-list">
      <h3>Your Orders</h3>
      {orders.length === 0 ? (
        <p>No orders found</p>
      ) : (
        <table>
          <thead>
            <tr>
              <th>ID</th>
              <th>Amount</th>
              <th>Description</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            {orders.map(order => (
              <tr key={order.id}>
                <td>{order.id.substring(0, 8)}...</td>
                <td>${order.amount.toFixed(2)}</td>
                <td>{order.description}</td>
                <td className={`status ${order.status.toLowerCase()}`}>
                  {order.status}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
};

export default OrderList;