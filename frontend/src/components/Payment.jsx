import React, { useState } from 'react';

const Payment = () => {
  const [userId, setUserId] = useState('test_user');
  const [orderId, setOrderId] = useState('');
  const [amount, setAmount] = useState('');
  const [paymentResult, setPaymentResult] = useState(null);
  const [message, setMessage] = useState('');
  const [loading, setLoading] = useState(false);

  const processPayment = async () => {
    if (!userId || !orderId || !amount) {
      setMessage('All fields are required');
      return;
    }
    
    setLoading(true);
    try {
      const response = await fetch('http://localhost:8000/payments/process-payment', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          user_id: userId,
          order_id: orderId,
          amount: parseFloat(amount)
        })
      });
      
      if (!response.ok) throw new Error('Payment processing failed');
      
      const data = await response.json();
      setPaymentResult(data);
      setMessage(data.success ? 'Payment successful!' : `Payment failed: ${data.message}`);
    } catch (error) {
      setMessage(error.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="payment-container">
      <h2>Payment Processing</h2>
      
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
          type="text"
          placeholder="Order ID"
          value={orderId}
          onChange={(e) => setOrderId(e.target.value)}
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
      
      <div className="action-buttons">
        <button onClick={processPayment} disabled={loading}>
          {loading ? 'Processing...' : 'Process Payment'}
        </button>
      </div>
      
      {paymentResult && (
        <div className="payment-result">
          <h3>Payment Result</h3>
          <p><strong>Success:</strong> {paymentResult.success.toString()}</p>
          <p><strong>Message:</strong> {paymentResult.message}</p>
          {paymentResult.order_id && <p><strong>Order ID:</strong> {paymentResult.order_id}</p>}
          {paymentResult.amount && <p><strong>Amount:</strong> ${paymentResult.amount}</p>}
        </div>
      )}
      
      {message && <div className={`message ${message.includes('failed') ? 'error' : ''}`}>
        {message}
      </div>}
    </div>
  );
};

export default Payment;