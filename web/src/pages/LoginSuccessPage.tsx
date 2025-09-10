import { useEffect } from 'react';
import { convertKeysToPascalCase } from '../services/api';
import { useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { Box, CircularProgress, Typography } from '@mui/material';
import type { User } from '../types';

const LoginSuccessPage = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const { setUser } = useAuth();

  useEffect(() => {
    const params = new URLSearchParams(location.search);
    const token = params.get('token');
    const refreshToken = params.get('refresh');
    const userBase64 = params.get('user');

    if (token && refreshToken && userBase64) {
      try {
        const userJSON = atob(userBase64);
        const user: User = convertKeysToPascalCase(JSON.parse(userJSON));

        localStorage.setItem('token', token);
        localStorage.setItem('refreshToken', refreshToken);
        localStorage.setItem('user', JSON.stringify(user));
        
        // Directly update the auth context state
        setUser(user);
        
        // Now that the state is updated, navigate
        navigate('/todos', { replace: true });

      } catch (e) {
        console.error("Failed to decode user data:", e);
        navigate('/login?error=auth-failed', { replace: true });
      }
    } else {
      // Handle error case where tokens or user data are missing
      navigate('/login?error=auth-failed', { replace: true });
    }
    // We only want this effect to run once on mount.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', height: '100vh' }}>
      <CircularProgress />
      <Typography sx={{ mt: 2 }}>Finalizing login...</Typography>
    </Box>
  );
};

export default LoginSuccessPage; 
