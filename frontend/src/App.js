import React, { useState } from 'react';
import Navbar from './components/Navbar';
import Account from './components/Account';
import Order from './components/Order';
import Payment from './components/Payment';
import './styles.css';

function App() {
  const [activeTab, setActiveTab] = useState('account');

  return (
    <div className="app-container">
      <Navbar activeTab={activeTab} setActiveTab={setActiveTab} />
      
      <div className="content">
        {activeTab === 'account' && <Account />}
        {activeTab === 'order' && <Order />}
        {activeTab === 'payment' && <Payment />}
      </div>
    </div>
  );
}

export default App;