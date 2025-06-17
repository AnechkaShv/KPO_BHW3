import React, { useState } from 'react';
import { 
  Button, 
  TextField, 
  Paper, 
  Typography, 
  Grid,
  Snackbar,
  Alert,
  CircularProgress
} from '@mui/material';
import { createOrder } from '../api/orderApi';

const OrderTab = () => {
  const [userId, setUserId] = useState('');
  const [amount, setAmount] = useState('');
  const [description, setDescription] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);

  const handleCreateOrder = async () => {
    try {
      setLoading(true);
      setError(null);
      
      const order = await createOrder(userId, amount, description);
      setSuccess(`Order created with ID: ${order.id}`);
      
      setUserId('');
      setAmount('');
      setDescription('');
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Paper elevation={3} sx={{ p: 3, position: 'relative' }}>
      {loading && <CircularProgress sx={{ position: 'absolute', top: 16, right: 16 }} />}
      
      <Typography variant="h5" gutterBottom>Create Order</Typography>
      
      <Grid container spacing={2}>
        <Grid item xs={12} sm={4}>
          <TextField
            fullWidth
            label="User ID"
            value={userId}
            onChange={(e) => setUserId(e.target.value)}
            disabled={loading}
          />
        </Grid>
        <Grid item xs={12} sm={4}>
          <TextField
            fullWidth
            label="Amount"
            type="number"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            disabled={loading}
          />
        </Grid>
        <Grid item xs={12} sm={4}>
          <TextField
            fullWidth
            label="Description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            disabled={loading}
          />
        </Grid>
        <Grid item xs={12}>
          <Button
            variant="contained"
            onClick={handleCreateOrder}
            disabled={loading || !userId || !amount || !description}
          >
            Create Order
          </Button>
        </Grid>
      </Grid>

      <Snackbar
        open={!!error}
        autoHideDuration={6000}
        onClose={() => setError(null)}
      >
        <Alert severity="error">{error}</Alert>
      </Snackbar>

      <Snackbar
        open={!!success}
        autoHideDuration={6000}
        onClose={() => setSuccess(null)}
      >
        <Alert severity="success">{success}</Alert>
      </Snackbar>
    </Paper>
  );
};

export default OrderTab;