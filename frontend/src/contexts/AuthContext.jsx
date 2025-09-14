import React, { useState, useEffect } from 'react';
import { userService, eventService, getUser, getAdmin, logout, adminLogout } from '../services/api';
import { AuthContext } from './auth';

export const AuthProvider = ({ children }) => {
    const [user, setUser] = useState(null);
    const [admin, setAdmin] = useState(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        // Check for existing user session
        const savedUser = getUser();
        const savedAdmin = getAdmin();

        if (savedUser) {
            setUser(savedUser);
        }

        if (savedAdmin) {
            setAdmin(savedAdmin);
        }

        setLoading(false);
    }, []);


    const userLogin = async (credentials) => {
        try {
            const response = await userService.login(credentials);
            const userData = {
                user_id: response.data.user_id,
                email: response.data.email,
                name: response.data.name
            };

            localStorage.setItem('access_token', response.data.access_token);
            localStorage.setItem('refresh_token', response.data.refresh_token);
            localStorage.setItem('user', JSON.stringify(userData));

            setUser(userData);
            return { success: true, data: response.data };
        } catch (error) {
            return { success: false, error: error.response?.data?.error || 'Login failed' };
        }
    };

    const userRegister = async (userData) => {
        try {
            const response = await userService.register(userData);
            const userInfo = {
                user_id: response.data.user_id,
                email: response.data.email,
                name: response.data.name
            };

            localStorage.setItem('access_token', response.data.access_token);
            localStorage.setItem('refresh_token', response.data.refresh_token);
            localStorage.setItem('user', JSON.stringify(userInfo));

            setUser(userInfo);
            return { success: true, data: response.data };
        } catch (error) {
            return { success: false, error: error.response?.data?.error || 'Registration failed' };
        }
    };

    const userLogout = async () => {
        try {
            await logout();
            setUser(null);
            return { success: true };
        } catch {
            setUser(null);
            return { success: true };
        }
    };

    const updateUserProfile = async (profileData) => {
        try {
            const response = await userService.updateProfile(profileData);
            const updatedUser = {
                user_id: response.data.user_id,
                email: response.data.email,
                name: response.data.name
            };

            localStorage.setItem('user', JSON.stringify(updatedUser));
            setUser(updatedUser);
            return { success: true, data: response.data };
        } catch (error) {
            return { success: false, error: error.response?.data?.error || 'Profile update failed' };
        }
    };


    const adminLogin = async (credentials) => {
        try {
            const response = await eventService.adminLogin(credentials);
            const adminData = {
                admin_id: response.data.admin_id,
                email: response.data.email,
                name: response.data.name,
                role: response.data.role
            };

            localStorage.setItem('admin_access_token', response.data.access_token);
            localStorage.setItem('admin_refresh_token', response.data.refresh_token);
            localStorage.setItem('admin', JSON.stringify(adminData));

            setAdmin(adminData);
            return { success: true, data: response.data };
        } catch (error) {
            return { success: false, error: error.response?.data?.error || 'Admin login failed' };
        }
    };

    const adminRegister = async (adminData) => {
        try {
            const response = await eventService.adminRegister(adminData);
            const adminInfo = {
                admin_id: response.data.admin_id,
                email: response.data.email,
                name: response.data.name,
                role: response.data.role
            };

            localStorage.setItem('admin_access_token', response.data.access_token);
            localStorage.setItem('admin_refresh_token', response.data.refresh_token);
            localStorage.setItem('admin', JSON.stringify(adminInfo));

            setAdmin(adminInfo);
            return { success: true, data: response.data };
        } catch (error) {
            return { success: false, error: error.response?.data?.error || 'Admin registration failed' };
        }
    };

    const adminLogoutHandler = () => {
        adminLogout();
        setAdmin(null);
    };

    // ==================== CONTEXT VALUE ====================

    const value = {
        // User state
        user,
        isAuthenticated: !!user,

        // Admin state
        admin,
        isAdminAuthenticated: !!admin,

        // Loading state
        loading,

        // User methods
        userLogin,
        userRegister,
        userLogout,
        updateUserProfile,

        // Admin methods
        adminLogin,
        adminRegister,
        adminLogout: adminLogoutHandler
    };

    return (
        <AuthContext.Provider value={value}>
            {children}
        </AuthContext.Provider>
    );
};

export { AuthProvider as default };
