import axios from 'axios';

// Base URLs for different services
const API_URLS = {
  USER_SERVICE: 'http://localhost:8001',
  EVENT_SERVICE: 'http://localhost:8002', 
  SEARCH_SERVICE: 'http://localhost:8003',
  BOOKING_SERVICE: 'http://localhost:8004'
};

// Create axios instances for each service
const userAPI = axios.create({
  baseURL: API_URLS.USER_SERVICE,
  headers: { 'Content-Type': 'application/json' }
});

const eventAPI = axios.create({
  baseURL: API_URLS.EVENT_SERVICE,
  headers: { 'Content-Type': 'application/json' }
});

const searchAPI = axios.create({
  baseURL: API_URLS.SEARCH_SERVICE,
  headers: { 'Content-Type': 'application/json' }
});

const bookingAPI = axios.create({
  baseURL: API_URLS.BOOKING_SERVICE,
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
            const response = await userAPI.post('/api/v1/auth/refresh', {
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

// Admin API interceptor setup
const adminAPI = axios.create({
  baseURL: API_URLS.EVENT_SERVICE,
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
          const response = await eventAPI.post('/api/v1/auth/admin/refresh', {
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
  register: (userData) => userAPI.post('/api/v1/auth/register', userData),
  login: (credentials) => userAPI.post('/api/v1/auth/login', credentials),
  logout: (refreshToken) => userAPI.post('/api/v1/auth/logout', { refresh_token: refreshToken }),
  refreshToken: (refreshToken) => userAPI.post('/api/v1/auth/refresh', { refresh_token: refreshToken }),
  
  // Profile
  getProfile: () => userAPI.get('/api/v1/users/profile'),
  updateProfile: (profileData) => userAPI.put('/api/v1/users/profile', profileData),
  getUserBookings: () => userAPI.get('/api/v1/users/bookings'),
  
  // Health
  health: () => userAPI.get('/healthz'),
  healthReady: () => userAPI.get('/health/ready')
};

// ==================== EVENT SERVICE APIs ====================

export const eventService = {
  // Admin Authentication
  adminRegister: (adminData) => eventAPI.post('/api/v1/auth/admin/register', adminData),
  adminLogin: (credentials) => eventAPI.post('/api/v1/auth/admin/login', credentials),
  adminRefresh: (refreshToken) => eventAPI.post('/api/v1/auth/admin/refresh', { refresh_token: refreshToken }),
  
  // Venues (Admin)
  createVenue: (venueData) => adminAPI.post('/api/v1/admin/venues', venueData),
  getVenues: (params = {}) => adminAPI.get('/api/v1/admin/venues', { params }),
  updateVenue: (venueId, venueData) => adminAPI.put(`/api/v1/admin/venues/${venueId}`, venueData),
  deleteVenue: (venueId) => adminAPI.delete(`/api/v1/admin/venues/${venueId}`),
  
  // Events (Admin)
  createEvent: (eventData) => adminAPI.post('/api/v1/admin/events', eventData),
  getAdminEvents: (params = {}) => adminAPI.get('/api/v1/admin/events', { params }),
  updateEvent: (eventId, eventData) => adminAPI.put(`/api/v1/admin/events/${eventId}`, eventData),
  deleteEvent: (eventId, version) => adminAPI.delete(`/api/v1/admin/events/${eventId}`, { data: { version } }),
  
  // Public Events
  getEvents: (params = {}) => eventAPI.get('/api/v1/events', { params }),
  getEvent: (eventId) => eventAPI.get(`/api/v1/events/${eventId}`),
  getEventAvailability: (eventId) => eventAPI.get(`/api/v1/events/${eventId}/availability`),
  
  // Health
  health: () => eventAPI.get('/healthz'),
  healthReady: () => eventAPI.get('/health/ready')
};

// ==================== SEARCH SERVICE APIs ====================

export const searchService = {
  // Search
  search: (params = {}) => searchAPI.get('/api/v1/search', { params }),
  getSuggestions: (params) => searchAPI.get('/api/v1/search/suggestions', { params }),
  getFilters: () => searchAPI.get('/api/v1/search/filters'),
  getTrending: (params = {}) => searchAPI.get('/api/v1/search/trending', { params }),
  
  // Health
  health: () => searchAPI.get('/healthz'),
  healthReady: () => searchAPI.get('/health/ready')
};

// ==================== BOOKING SERVICE APIs ====================

export const bookingService = {
  // Availability
  checkAvailability: (params) => bookingAPI.get('/api/v1/bookings/check-availability', { params }),
  
  // Booking Flow
  reserve: (bookingData) => bookingAPI.post('/api/v1/bookings/reserve', bookingData),
  confirm: (confirmData) => bookingAPI.post('/api/v1/bookings/confirm', confirmData),
  getBooking: (bookingId) => bookingAPI.get(`/api/v1/bookings/${bookingId}`),
  cancelBooking: (bookingId) => bookingAPI.delete(`/api/v1/bookings/${bookingId}`),
  expireReservation: (reservationId) => bookingAPI.post(`/api/v1/bookings/${reservationId}/expire`),
  getUserBookings: (userId, params = {}) => bookingAPI.get(`/api/v1/bookings/user/${userId}`, { params }),
  
  // Waitlist
  joinWaitlist: (waitlistData) => bookingAPI.post('/api/v1/waitlist/join', waitlistData),
  getWaitlistPosition: (params) => bookingAPI.get('/api/v1/waitlist/position', { params }),
  leaveWaitlist: (waitlistData) => bookingAPI.delete('/api/v1/waitlist/leave', { data: waitlistData }),
  
  // Health
  health: () => bookingAPI.get('/healthz'),
  healthReady: () => bookingAPI.get('/health/ready')
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
