import { useForm } from 'react-hook-form';
import { useAuth } from '../context/AuthContext';
import { useNavigate, Link as RouterLink } from 'react-router-dom';
import { Container, Box, TextField, Button, Typography, Alert, CircularProgress } from '@mui/material';
import { useState } from 'react';

const RegisterPage = () => {
  const { register: formRegister, handleSubmit, formState: { errors } } = useForm();
  const { register: authRegister } = useAuth();
  const navigate = useNavigate();
  const [serverError, setServerError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const onSubmit = async (data: any) => {
    setLoading(true);
    setServerError(null);
    try {
      await authRegister(data);
      navigate('/login');
    } catch (error: any) {
      setServerError(error.response?.data?.message || 'Registration failed. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Container maxWidth="xs">
      <Box sx={{ mt: 8, display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
        <Typography component="h1" variant="h5">
          Sign up
        </Typography>
        <Box component="form" onSubmit={handleSubmit(onSubmit)} sx={{ mt: 3 }}>
          <TextField
            margin="normal"
            required
            fullWidth
            id="firstName"
            label="First Name"
            autoFocus
            {...formRegister('FirstName', { required: 'First name is required' })}
            error={!!errors.FirstName}
            helperText={errors.FirstName?.message as string}
          />
          <TextField
            margin="normal"
            required
            fullWidth
            id="lastName"
            label="Last Name"
            {...formRegister('LastName', { required: 'Last name is required' })}
            error={!!errors.LastName}
            helperText={errors.LastName?.message as string}
          />
          <TextField
            margin="normal"
            required
            fullWidth
            id="email"
            label="Email Address"
            autoComplete="email"
            {...formRegister('Email', { required: 'Email is required' })}
            error={!!errors.Email}
            helperText={errors.Email?.message as string}
          />
          <TextField
            margin="normal"
            required
            fullWidth
            label="Password"
            type="password"
            id="password"
            {...formRegister('Password', { required: 'Password is required' })}
            error={!!errors.Password}
            helperText={errors.Password?.message as string}
          />
          {serverError && <Alert severity="error" sx={{ mt: 2 }}>{serverError}</Alert>}
          <Button
            type="submit"
            fullWidth
            variant="contained"
            sx={{ mt: 3, mb: 2 }}
            disabled={loading}
          >
            {loading ? <CircularProgress size={24} /> : 'Sign Up'}
          </Button>
          <Typography variant="body2" align="center">
            {'Already have an account? '}
            <RouterLink to="/login">
              Sign In
            </RouterLink>
          </Typography>
        </Box>
      </Box>
    </Container>
  );
};

export default RegisterPage; 