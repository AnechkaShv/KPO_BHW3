import React, { useState, useEffect } from 'react';
import {
  Button,
  TextField,
  Paper,
  Typography,
  Grid,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Snackbar,
  Alert,
  CircularProgress,
  Box
} from '@mui/material';
import { createAccount, getAccount, deposit } from '../api/paymentApi';

const AccountTab = () => {
  const [userId, setUserId] = useState('');
  const [amount, setAmount] = useState('');
  const [accounts, setAccounts] = useState([]);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);
  const [loading, setLoading] = useState(false);

  const handleCreateAccount = async () => {
    try {
      setLoading(true);
      setError(null);
      const account = await createAccount(userId);
      setSuccess(`Account created with ID: ${account.id}`);
      setAccounts([account]);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleGetAccount = async () => {
    try {
      setLoading(true);
      setError(null);
      const account = await getAccount(userId);
      setAccounts([account]);
      setSuccess('Account loaded successfully');
    } catch (err) {
      setError(err.message);
      setAccounts([]);
    } finally {
      setLoading(false);
    }
  };

  const handleDeposit = async () => {
    try {
      setLoading(true);
      setError(null);
      await deposit(userId, parseFloat(amount));
      setSuccess('Deposit successful');
      await handleGetAccount(); // Refresh account data
      setAmount('');
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleCloseSnackbar = () => {
    setError(null);
    setSuccess(null);
  };

  return (
    <Paper elevation={3} sx={{ p: 3, position: 'relative' }}>
      {loading && (
        <Box
          sx={{
            position: 'absolute',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            backgroundColor: 'rgba(255, 255, 255, 0.7)',
            zIndex: 1
          }}
        >
          <CircularProgress />
        </Box>
      )}

      <Typography variant="h5" gutterBottom>
        Account Management
      </Typography>

      <Grid container spacing={2} sx={{ mb: 3 }}>
        <Grid item xs={12} sm={6}>
          <TextField
            fullWidth
            label="User ID"
            variant="outlined"
            value={userId}
            onChange={(e) => setUserId(e.target.value)}
            disabled={loading}
          />
        </Grid>
        <Grid item xs={12} sm={6}>
          <Button
            variant="contained"
            color="primary"
            onClick={handleCreateAccount}
            sx={{ mr: 2 }}
            disabled={loading || !userId}
          >
            Create Account
          </Button>
          <Button
            variant="contained"
            color="secondary"
            onClick={handleGetAccount}
            disabled={loading || !userId}
          >
            Get Account
          </Button>
        </Grid>
      </Grid>

      <Grid container spacing={2} sx={{ mb: 3 }}>
        <Grid item xs={12} sm={6}>
          <TextField
            fullWidth
            label="Amount"
            variant="outlined"
            type="number"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            disabled={loading}
          />
        </Grid>
        <Grid item xs={12} sm={6}>
          <Button
            variant="contained"
            onClick={handleDeposit}
            disabled={loading || !userId || !amount}
          >
            Deposit
          </Button>
        </Grid>
      </Grid>

      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>ID</TableCell>
              <TableCell>User ID</TableCell>
              <TableCell>Balance</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {accounts.map((account) => (
              <TableRow key={account.id}>
                <TableCell>{account.id}</TableCell>
                <TableCell>{account.user_id}</TableCell>
                <TableCell>${account.balance?.toFixed(2)}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>

      <Snackbar
        open={!!error}
        autoHideDuration={6000}
        onClose={handleCloseSnackbar}
      >
        <Alert severity="error" onClose={handleCloseSnackbar}>
          {error}
        </Alert>
      </Snackbar>

      <Snackbar
        open={!!success}
        autoHideDuration={6000}
        onClose={handleCloseSnackbar}
      >
        <Alert severity="success" onClose={handleCloseSnackbar}>
          {success}
        </Alert>
      </Snackbar>
    </Paper>
  );
};

export default AccountTab;