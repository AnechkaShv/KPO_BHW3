const API_BASE_URL = process.env.REACT_APP_API_BASE_URL || 'http://localhost:8081';

const handleResponse = async (response) => {
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
  }
  return response.json();
};

// Создание аккаунта
export const createAccount = async (userId) => {
  const response = await fetch(`${API_BASE_URL}/api/payments/create-account`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ user_id: userId }),
  });
  return handleResponse(response);
};

// Получение информации об аккаунте
export const getAccount = async (userId) => {
  const response = await fetch(`${API_BASE_URL}/api/payments/get-account?user_id=${userId}`);
  return handleResponse(response);
};

// Пополнение счета
export const deposit = async (userId, amount) => {
  const response = await fetch(`${API_BASE_URL}/api/payments/deposit`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ user_id: userId, amount: parseFloat(amount) }),
  });
  return handleResponse(response);
};

// Обработка платежа
export const processPayment = async (orderId, userId, amount) => {
  const response = await fetch(`${API_BASE_URL}/api/payments/process`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      order_id: orderId,
      user_id: userId,
      amount: parseFloat(amount)
    }),
  });
  return handleResponse(response);
};