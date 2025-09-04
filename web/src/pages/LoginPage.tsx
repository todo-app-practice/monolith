import { useForm } from 'react-hook-form';
import { useAuth } from '../context/AuthContext';
import { useNavigate, Link as RouterLink } from 'react-router-dom';
import { Container, Box, TextField, Button, Typography, Alert, CircularProgress } from '@mui/material';
import { useState } from 'react';
import GoogleIcon from '@mui/icons-material/Google';

const LoginPage = () => {
  const { register, handleSubmit, formState: { errors } } = useForm();
  const { login } = useAuth();
  const navigate = useNavigate();
  const [serverError, setServerError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const onSubmit = async (data: any) => {
    setLoading(true);
    setServerError(null);
    try {
      const user = await login(data);
      if (user) {
        navigate('/todos');
      }
    } catch (error: any) {
      setServerError(error.response?.data?.message || 'Login failed. Please check your credentials.');
    } finally {
      setLoading(false);
    }
  };

  const handleGoogleLogin = () => {
    window.location.href = '/auth/google/login';
  };

  return (
    <Container maxWidth="xs">
      <Box sx={{ mt: 8, display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
        <Typography component="h1" variant="h5">
          Sign in
        </Typography>
        <Box component="form" onSubmit={handleSubmit(onSubmit)} sx={{ mt: 1 }}>
          <TextField
            margin="normal"
            required
            fullWidth
            id="email"
            label="Email Address"
            autoComplete="email"
            autoFocus
            {...register('email', { required: 'Email is required' })}
            error={!!errors.email}
            helperText={errors.email?.message as string}
          />
          <TextField
            margin="normal"
            required
            fullWidth
            label="Password"
            type="password"
            id="password"
            autoComplete="current-password"
            {...register('password', { required: 'Password is required' })}
            error={!!errors.password}
            helperText={errors.password?.message as string}
          />
          {serverError && <Alert severity="error" sx={{ mt: 2 }}>{serverError}</Alert>}
          <Button
            type="submit"
            fullWidth
            variant="contained"
            sx={{ mt: 3, mb: 2 }}
            disabled={loading}
          >
            {loading ? <CircularProgress size={24} /> : 'Sign In'}
          </Button>
          <Button
            component="a"
            href="/auth/google/login"
            fullWidth
            variant="outlined"
            sx={{ mt: 1, mb: 2 }}
            startIcon={<GoogleIcon />}
          >
            Sign In with Google
          </Button>
          <Typography variant="body2" align="center">
            {"Don't have an account? "}
            <RouterLink to="/register">
              Sign Up
            </RouterLink>
          </Typography>
        </Box>
      </Box>
    </Container>
  );
};

export default LoginPage; 