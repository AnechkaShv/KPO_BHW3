// frontend/src/components/CreateOrderForm.jsx
import React, { useState } from 'react';
import { createOrder } from '../api';

const CreateOrderForm = ({ onOrderCreated }) => {
  const [amount, setAmount] = useState('');
  const [description, setDescription] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState(null);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setIsSubmitting(true);
    setError(null);

    try {
      const order = await createOrder(
        'test_user', // Можно заменить на динамический user_id
        amount,
        description
      );
      onOrderCreated(order);
      setAmount('');
      setDescription('');
    } catch (err) {
      setError(err.message);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="create-order-form">
      <h2>Create New Order</h2>
      {error && <div className="error-message">{error}</div>}
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label>Amount:</label>
          <input
            type="number"
            step="0.01"
            min="0.01"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            required
          />
        </div>
        <div className="form-group">
          <label>Description:</label>
          <input
            type="text"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            required
          />
        </div>
        <button type="submit" disabled={isSubmitting}>
          {isSubmitting ? 'Creating...' : 'Create Order'}
        </button>
      </form>
    </div>
  );
};

export default CreateOrderForm;