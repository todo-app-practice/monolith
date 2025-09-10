import { useState, useEffect } from 'react';
import api from '../services/api';
import { useAuth } from '../context/AuthContext';
import type { Todo } from '../types';
import {
  Container, Box, TextField, Button, Typography, List, ListItem, ListItemText, Checkbox, IconButton, CircularProgress, Alert
} from '@mui/material';
import DeleteIcon from '@mui/icons-material/Delete';

const TodosPage = () => {
  const [todos, setTodos] = useState<Todo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [newTodoText, setNewTodoText] = useState('');
  const { user } = useAuth();

  const fetchTodos = async () => {
    try {
      setLoading(true);
      const { data } = await api.get('/todos');
      setTodos(data.Data || []);
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to fetch todos.');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTodos();
  }, []);

  const handleAddTodo = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newTodoText.trim()) return;
    if (!user) {
      setError('User not authenticated');
      return;
    }
    
    try {
      const { data } = await api.post('/todos', { 
        text: newTodoText,
        userId: user.Id
      });
      setTodos([...todos, data]);
      setNewTodoText('');
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to add todo.');
    }
  };

  const handleToggleTodo = async (id: number, done: boolean) => {
    try {
      await api.put(`/todos/${id}`, { done });
      setTodos(todos.map(todo => (todo.Id === id ? { ...todo, Done: done } : todo)));
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to update todo.');
    }
  };

  const handleDeleteTodo = async (id: number) => {
    try {
      await api.delete(`/todos/${id}`);
      setTodos(todos.filter(todo => todo.Id !== id));
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to delete todo.');
    }
  };

  const handleUpdateText = async (id: number, text: string) => {
    try {
      await api.put(`/todos/${id}`, { text });
      // We don't update the local state immediately to avoid flickering.
      // A full fetch could be an alternative.
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to update todo text.');
    }
  };


  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', mt: 4 }}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Container maxWidth="md">
      <Box sx={{ mt: 4 }}>
        <Typography variant="h4" component="h1" gutterBottom>
          My Todos
        </Typography>
        {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}
        <Box component="form" onSubmit={handleAddTodo} sx={{ display: 'flex', mb: 2 }}>
          <TextField
            fullWidth
            variant="outlined"
            label="New Todo"
            value={newTodoText}
            onChange={(e) => setNewTodoText(e.target.value)}
          />
          <Button type="submit" variant="contained" sx={{ ml: 2, whiteSpace: 'nowrap' }}>
            Add Todo
          </Button>
        </Box>
        <List>
          {todos.map((todo) => (
            <ListItem
              key={todo.Id}
              secondaryAction={
                <IconButton edge="end" aria-label="delete" onClick={() => handleDeleteTodo(todo.Id)}>
                  <DeleteIcon color="error" />
                </IconButton>
              }
            >
              <Checkbox
                edge="start"
                checked={todo.Done}
                onChange={() => handleToggleTodo(todo.Id, !todo.Done)}
              />
              <ListItemText
                primary={
                  <TextField
                    defaultValue={todo.Text}
                    variant="standard"
                    fullWidth
                    onBlur={(e) => handleUpdateText(todo.Id, e.target.value)}
                    sx={{
                      textDecoration: todo.Done ? 'line-through' : 'none',
                    }}
                  />
                }
              />
            </ListItem>
          ))}
        </List>
      </Box>
    </Container>
  );
};

export default TodosPage; 
