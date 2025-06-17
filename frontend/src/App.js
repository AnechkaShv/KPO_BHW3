import React, { useState } from 'react';
import { Container, Typography, Tabs, Tab, Box, CssBaseline } from '@mui/material';
import { ThemeProvider, createTheme } from '@mui/material/styles';
import AccountTab from './components/AccountTab';
import OrderTab from './components/OrderTab';
import PaymentTab from './components/PaymentTab';

const theme = createTheme({
  palette: {
    primary: {
      main: '#1976d2',
    },
    secondary: {
      main: '#dc004e',
    },
  },
});

function TabPanel(props) {
  const { children, value, index, ...other } = props;

  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`simple-tabpanel-${index}`}
      aria-labelledby={`simple-tab-${index}`}
      {...other}
    >
      {value === index && <Box sx={{ p: 3 }}>{children}</Box>}
    </div>
  );
}

function App() {
  const [tabValue, setTabValue] = useState(0);

  const handleChange = (event, newValue) => {
    setTabValue(newValue);
  };

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Container maxWidth="md" sx={{ mt: 4, mb: 4 }}>
        <Typography variant="h3" component="h1" gutterBottom align="center">
          Payment System Dashboard
        </Typography>

        <Tabs
          value={tabValue}
          onChange={handleChange}
          centered
          sx={{ mb: 2 }}
        >
          <Tab label="Accounts" />
          <Tab label="Orders" />
          <Tab label="Payments" />
        </Tabs>

        <TabPanel value={tabValue} index={0}>
          <AccountTab />
        </TabPanel>
        <TabPanel value={tabValue} index={1}>
          <OrderTab />
        </TabPanel>
        <TabPanel value={tabValue} index={2}>
          <PaymentTab />
        </TabPanel>
      </Container>
    </ThemeProvider>
  );
}

export default App;