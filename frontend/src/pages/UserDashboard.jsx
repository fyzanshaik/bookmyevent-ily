import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { bookingService, userService, formatError } from '../services/api';
import {
    User, Calendar, Ticket, Clock, CheckCircle,
    XCircle, AlertCircle, Edit, Trash2
} from 'lucide-react';

const UserDashboard = () => {
    const { user, updateUserProfile } = useAuth();
    const [bookings, setBookings] = useState([]);
    const [profile, setProfile] = useState(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [activeTab, setActiveTab] = useState('bookings');
    const [editingProfile, setEditingProfile] = useState(false);
    const [profileData, setProfileData] = useState({
        name: '',
        phone_number: ''
    });

    useEffect(() => {
        const fetchData = async () => {
            try {
                const [bookingsResponse, profileResponse] = await Promise.all([
                    bookingService.getUserBookings(user.user_id),
                    userService.getProfile()
                ]);

                setBookings(bookingsResponse.data.bookings || []);
                setProfile(profileResponse.data);
                setProfileData({
                    name: profileResponse.data.name,
                    phone_number: profileResponse.data.phone_number || ''
                });
            } catch (error) {
                setError(formatError(error));
            } finally {
                setLoading(false);
            }
        };

        if (user) {
            fetchData();
        }
    }, [user]);

    const handleCancelBooking = async (bookingId) => {
        if (!window.confirm('Are you sure you want to cancel this booking?')) {
            return;
        }

        try {
            await bookingService.cancelBooking(bookingId);
            // Refresh bookings
            const response = await bookingService.getUserBookings(user.user_id);
            setBookings(response.data.bookings || []);
        } catch (error) {
            setError(formatError(error));
        }
    };

    const handleProfileUpdate = async (e) => {
        e.preventDefault();

        try {
            const result = await updateUserProfile(profileData);
            if (result.success) {
                setProfile(result.data);
                setEditingProfile(false);
            } else {
                setError(result.error);
            }
        } catch (error) {
            setError(formatError(error));
        }
    };

    const getStatusIcon = (status) => {
        switch (status) {
            case 'confirmed':
                return <CheckCircle className="h-5 w-5 text-green-600" />;
            case 'pending':
                return <Clock className="h-5 w-5 text-yellow-600" />;
            case 'cancelled':
                return <XCircle className="h-5 w-5 text-red-600" />;
            case 'expired':
                return <AlertCircle className="h-5 w-5 text-gray-600" />;
            default:
                return <AlertCircle className="h-5 w-5 text-gray-600" />;
        }
    };

    const getStatusColor = (status) => {
        switch (status) {
            case 'confirmed':
                return 'bg-green-100 text-green-800';
            case 'pending':
                return 'bg-yellow-100 text-yellow-800';
            case 'cancelled':
                return 'bg-red-100 text-red-800';
            case 'expired':
                return 'bg-gray-100 text-gray-800';
            default:
                return 'bg-gray-100 text-gray-800';
        }
    };

    const formatDate = (dateString) => {
        return new Date(dateString).toLocaleDateString('en-US', {
            year: 'numeric',
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    };

    if (loading) {
        return (
            <div className="flex justify-center items-center min-h-[400px]">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
            </div>
        );
    }

    return (
        <div className="max-w-6xl mx-auto space-y-6">
            {/* Header */}
            <div className="bg-white rounded-lg shadow-md p-6">
                <div className="flex items-center space-x-4">
                    <div className="w-16 h-16 bg-blue-600 rounded-full flex items-center justify-center">
                        <User className="h-8 w-8 text-white" />
                    </div>
                    <div>
                        <h1 className="text-2xl font-bold text-gray-900">
                            Welcome back, {user?.name}!
                        </h1>
                        <p className="text-gray-600">Manage your bookings and profile</p>
                    </div>
                </div>
            </div>

            {/* Error Message */}
            {error && (
                <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded flex items-center">
                    <AlertCircle className="h-5 w-5 mr-2" />
                    {error}
                </div>
            )}

            {/* Tabs */}
            <div className="bg-white rounded-lg shadow-md">
                <div className="border-b border-gray-200">
                    <nav className="flex space-x-8 px-6">
                        <button
                            onClick={() => setActiveTab('bookings')}
                            className={`py-4 px-1 border-b-2 font-medium text-sm ${activeTab === 'bookings'
                                ? 'border-blue-500 text-blue-600'
                                : 'border-transparent text-gray-500 hover:text-gray-700'
                                }`}
                        >
                            <Ticket className="h-5 w-5 inline mr-2" />
                            My Bookings ({bookings.length})
                        </button>
                        <button
                            onClick={() => setActiveTab('profile')}
                            className={`py-4 px-1 border-b-2 font-medium text-sm ${activeTab === 'profile'
                                ? 'border-blue-500 text-blue-600'
                                : 'border-transparent text-gray-500 hover:text-gray-700'
                                }`}
                        >
                            <User className="h-5 w-5 inline mr-2" />
                            Profile
                        </button>
                    </nav>
                </div>

                <div className="p-6">
                    {/* Bookings Tab */}
                    {activeTab === 'bookings' && (
                        <div className="space-y-4">
                            {bookings.length > 0 ? (
                                bookings.map((booking) => (
                                    <div key={booking.booking_id} className="border border-gray-200 rounded-lg p-6">
                                        <div className="flex justify-between items-start mb-4">
                                            <div>
                                                <h3 className="text-lg font-semibold text-gray-900">
                                                    {booking.event_name || 'Event'}
                                                </h3>
                                                <p className="text-sm text-gray-600">
                                                    Booking Reference: {booking.booking_reference}
                                                </p>
                                            </div>

                                            <div className="flex items-center space-x-2">
                                                {getStatusIcon(booking.status)}
                                                <span className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(booking.status)}`}>
                                                    {booking.status}
                                                </span>
                                            </div>
                                        </div>

                                        <div className="grid md:grid-cols-3 gap-4 mb-4">
                                            <div>
                                                <p className="text-sm text-gray-600">Quantity</p>
                                                <p className="font-medium">{booking.quantity} ticket{booking.quantity > 1 ? 's' : ''}</p>
                                            </div>

                                            <div>
                                                <p className="text-sm text-gray-600">Total Amount</p>
                                                <p className="font-medium">${booking.total_amount}</p>
                                            </div>

                                            <div>
                                                <p className="text-sm text-gray-600">Booked On</p>
                                                <p className="font-medium">{formatDate(booking.booked_at)}</p>
                                            </div>
                                        </div>

                                        {booking.venue && (
                                            <div className="mb-4">
                                                <p className="text-sm text-gray-600">Venue</p>
                                                <p className="font-medium">{booking.venue}</p>
                                            </div>
                                        )}

                                        {booking.datetime && (
                                            <div className="mb-4">
                                                <div className="flex items-center text-gray-600">
                                                    <Calendar className="h-4 w-4 mr-2" />
                                                    <span>{formatDate(booking.datetime)}</span>
                                                </div>
                                            </div>
                                        )}

                                        <div className="flex justify-between items-center pt-4 border-t border-gray-200">
                                            <div className="flex space-x-3">
                                                {booking.status === 'confirmed' && booking.ticket_url && (
                                                    <a
                                                        href={booking.ticket_url}
                                                        target="_blank"
                                                        rel="noopener noreferrer"
                                                        className="text-blue-600 hover:text-blue-700 text-sm font-medium"
                                                    >
                                                        View Tickets
                                                    </a>
                                                )}

                                                {booking.status === 'pending' && booking.expires_at && (
                                                    <span className="text-yellow-600 text-sm">
                                                        Expires: {formatDate(booking.expires_at)}
                                                    </span>
                                                )}
                                            </div>

                                            {booking.status === 'confirmed' && (
                                                <button
                                                    onClick={() => handleCancelBooking(booking.booking_id)}
                                                    className="text-red-600 hover:text-red-700 text-sm font-medium flex items-center"
                                                >
                                                    <Trash2 className="h-4 w-4 mr-1" />
                                                    Cancel Booking
                                                </button>
                                            )}
                                        </div>
                                    </div>
                                ))
                            ) : (
                                <div className="text-center py-12">
                                    <Ticket className="h-16 w-16 text-gray-400 mx-auto mb-4" />
                                    <h3 className="text-lg font-semibold text-gray-900 mb-2">
                                        No bookings yet
                                    </h3>
                                    <p className="text-gray-600 mb-6">
                                        Start exploring events and book your first ticket!
                                    </p>
                                    <Link
                                        to="/events"
                                        className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 transition-colors"
                                    >
                                        Browse Events
                                    </Link>
                                </div>
                            )}
                        </div>
                    )}

                    {/* Profile Tab */}
                    {activeTab === 'profile' && profile && (
                        <div className="max-w-md">
                            {editingProfile ? (
                                <form onSubmit={handleProfileUpdate} className="space-y-4">
                                    <h3 className="text-lg font-semibold mb-4">Edit Profile</h3>

                                    <div>
                                        <label className="block text-sm font-medium text-gray-700 mb-1">
                                            Full Name
                                        </label>
                                        <input
                                            type="text"
                                            value={profileData.name}
                                            onChange={(e) => setProfileData({ ...profileData, name: e.target.value })}
                                            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                                            required
                                        />
                                    </div>

                                    <div>
                                        <label className="block text-sm font-medium text-gray-700 mb-1">
                                            Phone Number
                                        </label>
                                        <input
                                            type="tel"
                                            value={profileData.phone_number}
                                            onChange={(e) => setProfileData({ ...profileData, phone_number: e.target.value })}
                                            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                                            placeholder="+1234567890"
                                        />
                                    </div>

                                    <div className="flex space-x-3">
                                        <button
                                            type="submit"
                                            className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 transition-colors"
                                        >
                                            Save Changes
                                        </button>
                                        <button
                                            type="button"
                                            onClick={() => setEditingProfile(false)}
                                            className="border border-gray-300 text-gray-700 px-4 py-2 rounded-md hover:bg-gray-50 transition-colors"
                                        >
                                            Cancel
                                        </button>
                                    </div>
                                </form>
                            ) : (
                                <div className="space-y-4">
                                    <div className="flex justify-between items-center">
                                        <h3 className="text-lg font-semibold">Profile Information</h3>
                                        <button
                                            onClick={() => setEditingProfile(true)}
                                            className="text-blue-600 hover:text-blue-700 flex items-center text-sm"
                                        >
                                            <Edit className="h-4 w-4 mr-1" />
                                            Edit
                                        </button>
                                    </div>

                                    <div className="space-y-3">
                                        <div>
                                            <p className="text-sm text-gray-600">Email</p>
                                            <p className="font-medium">{profile.email}</p>
                                        </div>

                                        <div>
                                            <p className="text-sm text-gray-600">Full Name</p>
                                            <p className="font-medium">{profile.name}</p>
                                        </div>

                                        <div>
                                            <p className="text-sm text-gray-600">Phone Number</p>
                                            <p className="font-medium">{profile.phone_number || 'Not provided'}</p>
                                        </div>

                                        <div>
                                            <p className="text-sm text-gray-600">Member Since</p>
                                            <p className="font-medium">{formatDate(profile.created_at)}</p>
                                        </div>
                                    </div>
                                </div>
                            )}
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
};

export default UserDashboard;
