const API_BASE_URL = process.env.REACT_APP_API_BASE_URL || 'http://localhost:8080';

const handleResponse = async (response) => {
  const text = await response.text();
  if (!text) {
    throw new Error('Empty response from server');
  }

  try {
    const data = JSON.parse(text);
    if (!response.ok) {
      throw new Error(data.error || `HTTP error! status: ${response.status}`);
    }
    return data;
  } catch (e) {
    console.error('Failed to parse JSON:', text);
    throw new Error('Invalid JSON response from server');
  }
};

export const createOrder = async (userId, amount, description) => {
  const response = await fetch(`${API_BASE_URL}/api/orders/create`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ 
      user_id: userId,
      amount: parseFloat(amount),
      description
    }),
  });
  return handleResponse(response);
};