// frontend/src/api.js
const getApiBase = () => {
    // Для разработки вне Docker
    if (window.location.hostname === 'localhost') {
      return 'http://localhost:8000';
    }
    // Для продакшена в Docker
    return 'http://api-gateway:8000';
  };
  
  const API_BASE = getApiBase();
  
  const fetchWithTimeout = (url, options = {}, timeout = 5000) => {
    return Promise.race([
      fetch(url, options),
      new Promise((_, reject) =>
        setTimeout(() => reject(new Error('Request timeout')), timeout)
      )
    ]);
  };
  
  export const createOrder = async (userId, amount, description) => {
    try {
      console.log(`Creating order via ${API_BASE}`);
      const response = await fetchWithTimeout(`${API_BASE}/orders/create`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          user_id: userId,
          amount: parseFloat(amount),
          description
        })
      });
  
      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || `HTTP error! status: ${response.status}`);
      }
  
      return await response.json();
    } catch (error) {
      console.error('Network error:', error);
      throw new Error(`Failed to create order: ${error.message}`);
    }
  };
  
  export const getOrders = async (userId) => {
    try {
      console.log(`Fetching orders via ${API_BASE}`);
      const response = await fetchWithTimeout(`${API_BASE}/orders/list?user_id=${userId}`);
  
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
  
      const data = await response.json();
      return Array.isArray(data) ? data : [];
    } catch (error) {
      console.error('Network error:', error);
      throw new Error(`Failed to fetch orders: ${error.message}`);
    }
  };