import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import AuthProvider from './contexts/AuthContext';
import Navbar from './components/layout/Navbar';
import Home from './pages/Home';
import { UserLogin } from './pages/auth/UserLogin';
import UserRegister from './pages/auth/UserRegister';
import { AdminLogin } from './pages/auth/AdminLogin';
import AdminRegister from './pages/auth/AdminRegister';
import Events from './pages/Events';
import EventDetails from './pages/EventDetails';
import BookingFlow from './pages/BookingFlow';
import WaitlistPage from './pages/WaitlistPage';
import UserDashboard from './pages/UserDashboard';
import AdminDashboard from './pages/admin/AdminDashboard';
import AdminVenues from './pages/admin/AdminVenues';
import AdminEvents from './pages/admin/AdminEvents';
import ProtectedRoute from './components/auth/ProtectedRoute';
import AdminProtectedRoute from './components/auth/AdminProtectedRoute';

function App() {
  return (
    <AuthProvider>
      <Router>
        <div className="min-h-screen bg-gray-50">
          <Navbar />
          <main className="container mx-auto px-4 py-8">
            <Routes>
              {/* Public Routes */}
              <Route path="/" element={<Home />} />
              <Route path="/events" element={<Events />} />
              <Route path="/events/:eventId" element={<EventDetails />} />
              <Route path="/waitlist/:eventId" element={<WaitlistPage />} />

              {/* User Auth Routes */}
              <Route path="/login" element={<UserLogin />} />
              <Route path="/register" element={<UserRegister />} />

              {/* Admin Auth Routes */}
              <Route path="/admin/login" element={<AdminLogin />} />
              <Route path="/admin/register" element={<AdminRegister />} />

              {/* Protected User Routes */}
              <Route path="/dashboard" element={
                <ProtectedRoute>
                  <UserDashboard />
                </ProtectedRoute>
              } />
              <Route path="/book/:eventId" element={
                <ProtectedRoute>
                  <BookingFlow />
                </ProtectedRoute>
              } />

              {/* Protected Admin Routes */}
              <Route path="/admin/dashboard" element={
                <AdminProtectedRoute>
                  <AdminDashboard />
                </AdminProtectedRoute>
              } />
              <Route path="/admin/venues" element={
                <AdminProtectedRoute>
                  <AdminVenues />
                </AdminProtectedRoute>
              } />
              <Route path="/admin/events" element={
                <AdminProtectedRoute>
                  <AdminEvents />
                </AdminProtectedRoute>
              } />

              {/* Catch all route */}
              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
          </main>
        </div>
      </Router>
    </AuthProvider>
  );
}

export default App;
