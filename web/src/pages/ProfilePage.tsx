import { useForm } from 'react-hook-form';
import { useAuth } from '../context/AuthContext';
import { Container, Box, TextField, Button, Typography, Alert, CircularProgress } from '@mui/material';
import { useState, useEffect } from 'react';
import api from '../services/api';
import type { User } from '../types';

const ProfilePage = () => {
  const { user, fetchUser } = useAuth();
  
  const { 
    register, 
    handleSubmit, 
    reset, 
    formState: { errors, isDirty } 
  } = useForm<User>({
    defaultValues: user || undefined
  });

  const [serverError, setServerError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    // When the user context is loaded or updated, reset the form with the new data.
    if (user) {
      reset(user);
    }
  }, [user, reset]);

  const onSubmit = async (data: User) => {
    setLoading(true);
    setServerError(null);
    setSuccess(null);
    try {
      const { data: updatedUser } = await api.put(`/user/${user?.Id}`, data);
      // Directly update local storage and the auth context with the new user data from the response.
      localStorage.setItem('user', JSON.stringify(updatedUser));
      await fetchUser(); // This will now re-sync the context from the updated local storage.
      reset(updatedUser); // Reset the form to its new default state (not dirty)
      setSuccess('Profile updated successfully!');
    } catch (error: any) {
      setServerError(error.response?.data?.message || 'Failed to update profile.');
    } finally {
      setLoading(false);
    }
  };

  if (!user) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', mt: 4 }}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Container maxWidth="xs">
      <Box sx={{ mt: 8, display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
        <Typography component="h1" variant="h5">
          My Profile
        </Typography>
        <Box component="form" onSubmit={handleSubmit(onSubmit)} sx={{ mt: 3 }}>
          <TextField
            margin="normal"
            required
            fullWidth
            label="First Name"
            autoFocus
            {...register('FirstName', { required: 'First name is required' })}
            error={!!errors.FirstName}
            helperText={errors.FirstName?.message as string}
          />
          <TextField
            margin="normal"
            required
            fullWidth
            label="Last Name"
            {...register('LastName', { required: 'Last name is required' })}
            error={!!errors.LastName}
            helperText={errors.LastName?.message as string}
          />
          <TextField
            margin="normal"
            required
            fullWidth
            label="Email Address"
            autoComplete="email"
            {...register('Email', { required: 'Email is required' })}
            error={!!errors.Email}
            helperText={errors.Email?.message as string}
          />
          {serverError && <Alert severity="error" sx={{ mt: 2, width: '100%' }}>{serverError}</Alert>}
          {success && <Alert severity="success" sx={{ mt: 2, width: '100%' }}>{success}</Alert>}
          <Button
            type="submit"
            fullWidth
            variant="contained"
            sx={{ mt: 3, mb: 2 }}
            disabled={!isDirty || loading}
          >
            {loading ? <CircularProgress size={24} /> : 'Save Changes'}
          </Button>
        </Box>
      </Box>
    </Container>
  );
};

export default ProfilePage; 