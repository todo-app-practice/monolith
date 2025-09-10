import axios from 'axios';
import { camelCase, mapKeys, isObject, isArray, map, mapValues } from 'lodash';

// A truly recursive function to convert all keys in an object or array of objects.
export const convertKeysToPascalCase = (data: any): any => {
  if (isArray(data)) {
    return map(data, convertKeysToPascalCase);
  }
  if (isObject(data) && data !== null) {
    const newObject = mapKeys(data, (_value, key) => {
      const camel = camelCase(key);
      return camel.charAt(0).toUpperCase() + camel.slice(1);
    });

    return mapValues(newObject, (value) => {
      return convertKeysToPascalCase(value);
    });
  }
  return data;
};

const api = axios.create({
  baseURL: '/api', // This will be proxied by Vite to the backend
});

// Add a request interceptor to include the token in headers
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Add a response interceptor to transform snake_case to PascalCase
api.interceptors.response.use(
  (response) => {
    if (response.data) {
      response.data = convertKeysToPascalCase(response.data);
    }
    return response;
  },
  (error) => {
    return Promise.reject(error);
  }
);

export default api; 