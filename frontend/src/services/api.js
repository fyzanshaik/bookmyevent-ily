import axios from 'axios';

// API Gateway Base URL - All requests go through nginx gateway
const API_GATEWAY_URL = import.meta.env.VITE_API_URL || 'http://localhost';

// Gateway routes for different services
const API_ROUTES = {
  USER: '/api/user',
  EVENT: '/api/event',
  SEARCH: '/api/search',
  BOOKING: '/api/booking'
};

// Create axios instances for each service (all through gateway)
const userAPI = axios.create({
  baseURL: `${API_GATEWAY_URL}${API_ROUTES.USER}`,
  headers: { 'Content-Type': 'application/json' }
});

const eventAPI = axios.create({
  baseURL: `${API_GATEWAY_URL}${API_ROUTES.EVENT}`,
  headers: { 'Content-Type': 'application/json' }
});

const searchAPI = axios.create({
  baseURL: `${API_GATEWAY_URL}${API_ROUTES.SEARCH}`,
  headers: { 'Content-Type': 'application/json' }
});

const bookingAPI = axios.create({
  baseURL: `${API_GATEWAY_URL}${API_ROUTES.BOOKING}`,
  headers: { 'Content-Type': 'application/json' }
});

// Request interceptors to add auth tokens
const addAuthInterceptor = (apiInstance) => {
  apiInstance.interceptors.request.use((config) => {
    const token = localStorage.getItem('access_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  });
};

// Add auth interceptors to all APIs
addAuthInterceptor(userAPI);
addAuthInterceptor(eventAPI);
addAuthInterceptor(bookingAPI);

// Response interceptor for token refresh
const addResponseInterceptor = (apiInstance) => {
  apiInstance.interceptors.response.use(
    (response) => response,
    async (error) => {
      if (error.response?.status === 401) {
        // Try to refresh token
        const refreshToken = localStorage.getItem('refresh_token');
        if (refreshToken) {
          try {
            const response = await userAPI.post('/auth/refresh', {
              refresh_token: refreshToken
            });
            
            localStorage.setItem('access_token', response.data.access_token);
            localStorage.setItem('refresh_token', response.data.refresh_token);
            
            // Retry original request
            error.config.headers.Authorization = `Bearer ${response.data.access_token}`;
            return axios.request(error.config);
        } catch {
          // Refresh failed, logout user
          localStorage.removeItem('access_token');
          localStorage.removeItem('refresh_token');
          localStorage.removeItem('user');
          window.location.href = '/login';
        }
        }
      }
      return Promise.reject(error);
    }
  );
};

addResponseInterceptor(userAPI);
addResponseInterceptor(eventAPI);
addResponseInterceptor(bookingAPI);

// Admin API interceptor setup (also through gateway)
const adminAPI = axios.create({
  baseURL: `${API_GATEWAY_URL}${API_ROUTES.EVENT}`,
  headers: { 'Content-Type': 'application/json' }
});

adminAPI.interceptors.request.use((config) => {
  const adminToken = localStorage.getItem('admin_access_token');
  if (adminToken) {
    config.headers.Authorization = `Bearer ${adminToken}`;
  }
  return config;
});

// Admin response interceptor for token refresh
adminAPI.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401) {
      // Try to refresh admin token
      const refreshToken = localStorage.getItem('admin_refresh_token');
      if (refreshToken) {
        try {
          const response = await eventAPI.post('/auth/admin/refresh', {
            refresh_token: refreshToken
          });
          
          localStorage.setItem('admin_access_token', response.data.access_token);
          localStorage.setItem('admin_refresh_token', response.data.refresh_token);
          
          // Retry original request
          error.config.headers.Authorization = `Bearer ${response.data.access_token}`;
          return axios.request(error.config);
        } catch {
          // Refresh failed, logout admin
          localStorage.removeItem('admin_access_token');
          localStorage.removeItem('admin_refresh_token');
          localStorage.removeItem('admin');
          window.location.href = '/admin/login';
        }
      }
    }
    return Promise.reject(error);
  }
);

// ==================== USER SERVICE APIs ====================

export const userService = {
  // Authentication
  register: (userData) => userAPI.post('/auth/register', userData),
  login: (credentials) => userAPI.post('/auth/login', credentials),
  logout: (refreshToken) => userAPI.post('/auth/logout', { refresh_token: refreshToken }),
  refreshToken: (refreshToken) => userAPI.post('/auth/refresh', { refresh_token: refreshToken }),

  // Profile
  getProfile: () => userAPI.get('/users/profile'),
  updateProfile: (profileData) => userAPI.put('/users/profile', profileData),
  getUserBookings: () => userAPI.get('/users/bookings'),

  // Health (via gateway)
  health: () => userAPI.get('/healthz')
};

// ==================== EVENT SERVICE APIs ====================

export const eventService = {
  // Admin Authentication
  adminRegister: (adminData) => eventAPI.post('/auth/admin/register', adminData),
  adminLogin: (credentials) => eventAPI.post('/auth/admin/login', credentials),
  adminRefresh: (refreshToken) => eventAPI.post('/auth/admin/refresh', { refresh_token: refreshToken }),

  // Venues (Admin)
  createVenue: (venueData) => adminAPI.post('/admin/venues', venueData),
  getVenues: (params = {}) => adminAPI.get('/admin/venues', { params }),
  updateVenue: (venueId, venueData) => adminAPI.put(`/admin/venues/${venueId}`, venueData),
  deleteVenue: (venueId) => adminAPI.delete(`/admin/venues/${venueId}`),

  // Events (Admin)
  createEvent: (eventData) => adminAPI.post('/admin/events', eventData),
  getAdminEvents: (params = {}) => adminAPI.get('/admin/events', { params }),
  updateEvent: (eventId, eventData) => adminAPI.put(`/admin/events/${eventId}`, eventData),
  deleteEvent: (eventId, version) => adminAPI.delete(`/admin/events/${eventId}`, { data: { version } }),

  // Public Events
  getEvents: (params = {}) => eventAPI.get('/events', { params }),
  getEvent: (eventId) => eventAPI.get(`/events/${eventId}`),
  getEventAvailability: (eventId) => eventAPI.get(`/events/${eventId}/availability`),

  // Health (via gateway)
  health: () => eventAPI.get('/healthz')
};

// ==================== SEARCH SERVICE APIs ====================

export const searchService = {
  // Search
  search: (params = {}) => searchAPI.get('/search', { params }),
  getSuggestions: (params) => searchAPI.get('/search/suggestions', { params }),
  getFilters: () => searchAPI.get('/search/filters'),
  getTrending: (params = {}) => searchAPI.get('/search', { params }), // Use regular search for trending

  // Health (via gateway)
  health: () => searchAPI.get('/healthz')
};

// ==================== BOOKING SERVICE APIs ====================

export const bookingService = {
  // Availability
  checkAvailability: (params) => bookingAPI.get('/bookings/check-availability', { params }),

  // Booking Flow
  reserve: (bookingData) => bookingAPI.post('/bookings/reserve', bookingData),
  confirm: (confirmData) => bookingAPI.post('/bookings/confirm', confirmData),
  getBooking: (bookingId) => bookingAPI.get(`/bookings/${bookingId}`),
  cancelBooking: (bookingId) => bookingAPI.delete(`/bookings/${bookingId}`),
  expireReservation: (reservationId) => bookingAPI.post(`/bookings/${reservationId}/expire`),
  getUserBookings: (userId, params = {}) => bookingAPI.get(`/bookings/user/${userId}`, { params }),

  // Waitlist
  joinWaitlist: (waitlistData) => bookingAPI.post('/waitlist/join', waitlistData),
  getWaitlistPosition: (params) => bookingAPI.get('/waitlist/position', { params }),
  leaveWaitlist: (waitlistData) => bookingAPI.delete('/waitlist/leave', { data: waitlistData }),

  // Health (via gateway)
  health: () => bookingAPI.get('/healthz')
};

// ==================== UTILITY FUNCTIONS ====================

export const formatError = (error) => {
  if (error.response?.data?.error) {
    return error.response.data.error;
  }
  if (error.response?.data?.message) {
    return error.response.data.message;
  }
  return error.message || 'An unexpected error occurred';
};

export const isAuthenticated = () => {
  return !!localStorage.getItem('access_token');
};

export const isAdminAuthenticated = () => {
  const adminToken = localStorage.getItem('admin_access_token');
  return !!adminToken;
};

export const getUser = () => {
  const userStr = localStorage.getItem('user');
  return userStr ? JSON.parse(userStr) : null;
};

export const getAdmin = () => {
  const adminStr = localStorage.getItem('admin');
  return adminStr ? JSON.parse(adminStr) : null;
};

export const logout = async () => {
  const refreshToken = localStorage.getItem('refresh_token');
  if (refreshToken) {
    try {
      await userService.logout(refreshToken);
    } catch (error) {
      console.error('Logout error:', error);
    }
  }
  
  localStorage.removeItem('access_token');
  localStorage.removeItem('refresh_token');
  localStorage.removeItem('user');
};

export const adminLogout = () => {
  localStorage.removeItem('admin_access_token');
  localStorage.removeItem('admin_refresh_token');
  localStorage.removeItem('admin');
};

// Export admin API for admin-specific calls
export { adminAPI };
