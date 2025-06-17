import React from 'react';

const Navbar = ({ activeTab, setActiveTab }) => {
  return (
    <nav className="navbar">
      <button 
        className={activeTab === 'account' ? 'active' : ''}
        onClick={() => setActiveTab('account')}
      >
        Account
      </button>
      <button 
        className={activeTab === 'order' ? 'active' : ''}
        onClick={() => setActiveTab('order')}
      >
        Orders
      </button>
      <button 
        className={activeTab === 'payment' ? 'active' : ''}
        onClick={() => setActiveTab('payment')}
      >
        Payments
      </button>
    </nav>
  );
};

export default Navbar;