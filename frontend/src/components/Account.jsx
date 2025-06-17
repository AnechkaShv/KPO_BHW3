import React, { useState } from 'react';

const Account = () => {
  const [userId, setUserId] = useState('test_user');
  const [accountData, setAccountData] = useState(null);
  const [depositAmount, setDepositAmount] = useState('');
  const [message, setMessage] = useState('');
  const [loading, setLoading] = useState(false);

  const createAccount = async () => {
    setLoading(true);
    try {
      const response = await fetch('http://localhost:8000/payments/create-account', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ user_id: userId })
      });
      
      if (!response.ok) throw new Error('Failed to create account');
      
      const data = await response.json();
      setAccountData(data);
      setMessage('Account created successfully!');
    } catch (error) {
      setMessage(error.message);
    } finally {
      setLoading(false);
    }
  };

  const getAccount = async () => {
    if (!userId) {
      setMessage('User ID is required');
      return;
    }
    
    setLoading(true);
    try {
      const response = await fetch(`http://localhost:8000/payments/get-account?user_id=${userId}`);
      
      if (!response.ok) throw new Error('Failed to fetch account');
      
      const data = await response.json();
      setAccountData(data);
      setMessage('');
    } catch (error) {
      setMessage(error.message);
    } finally {
      setLoading(false);
    }
  };

  const deposit = async () => {
    if (!userId || !depositAmount) {
      setMessage('User ID and amount are required');
      return;
    }
    
    setLoading(true);
    try {
      const response = await fetch('http://localhost:8000/payments/deposit', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          user_id: userId,
          amount: parseFloat(depositAmount)
        })
      });
      
      if (!response.ok) throw new Error('Deposit failed');
      
      setMessage('Deposit successful!');
      getAccount(); // Refresh account data
    } catch (error) {
      setMessage(error.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="account-container">
      <h2>Account Management</h2>
      
      <div className="form-group">
        <input
          type="text"
          placeholder="User ID"
          value={userId}
          onChange={(e) => setUserId(e.target.value)}
        />
      </div>
      
      <div className="action-buttons">
        <button onClick={createAccount} disabled={loading}>
          {loading ? 'Processing...' : 'Create Account'}
        </button>
        <button onClick={getAccount} disabled={loading}>
          {loading ? 'Processing...' : 'Get Account'}
        </button>
      </div>
      
      {accountData && (
        <div className="account-info">
          <p><strong>User ID:</strong> {accountData.user_id}</p>
          <p><strong>Balance:</strong> ${accountData.balance}</p>
        </div>
      )}
      
      <div className="form-group">
        <input
          type="number"
          placeholder="Amount to deposit"
          value={depositAmount}
          onChange={(e) => setDepositAmount(e.target.value)}
        />
        <button onClick={deposit} disabled={loading}>
          {loading ? 'Processing...' : 'Deposit'}
        </button>
      </div>
      
      {message && <div className={`message ${message.includes('failed') ? 'error' : ''}`}>
        {message}
      </div>}
    </div>
  );
};

export default Account;