import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';
import { eventService, formatError } from '../../services/api';
import {
    Shield, Building, Calendar, Users,
    TrendingUp, Plus, Eye, Settings
} from 'lucide-react';

const AdminDashboard = () => {
    const { admin } = useAuth();
    const [stats, setStats] = useState({
        totalEvents: 0,
        totalVenues: 0,
        publishedEvents: 0,
        draftEvents: 0
    });
    const [recentEvents, setRecentEvents] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');

    useEffect(() => {
        const fetchDashboardData = async () => {
            try {
                // Fetch admin events
                const eventsResponse = await eventService.getAdminEvents({ limit: 10 });
                const events = eventsResponse.data.events || [];

                // Fetch venues
                const venuesResponse = await eventService.getVenues({ limit: 5 });
                const venues = venuesResponse.data.venues || [];

                setRecentEvents(events);
                setStats({
                    totalEvents: events.length,
                    totalVenues: venues.length,
                    publishedEvents: events.filter(e => e.status === 'published').length,
                    draftEvents: events.filter(e => e.status === 'draft').length
                });
            } catch (error) {
                setError(formatError(error));
            } finally {
                setLoading(false);
            }
        };

        fetchDashboardData();
    }, []);

    const formatDate = (dateString) => {
        return new Date(dateString).toLocaleDateString('en-US', {
            year: 'numeric',
            month: 'short',
            day: 'numeric'
        });
    };

    const getStatusColor = (status) => {
        switch (status) {
            case 'published':
                return 'bg-green-100 text-green-800';
            case 'draft':
                return 'bg-yellow-100 text-yellow-800';
            case 'cancelled':
                return 'bg-red-100 text-red-800';
            default:
                return 'bg-gray-100 text-gray-800';
        }
    };

    if (loading) {
        return (
            <div className="flex justify-center items-center min-h-[400px]">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-purple-600"></div>
            </div>
        );
    }

    return (
        <div className="max-w-7xl mx-auto space-y-6">
            {/* Header */}
            <div className="bg-white rounded-lg shadow-md p-6">
                <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-4">
                        <div className="w-16 h-16 bg-purple-600 rounded-full flex items-center justify-center">
                            <Shield className="h-8 w-8 text-white" />
                        </div>
                        <div>
                            <h1 className="text-2xl font-bold text-gray-900">
                                Admin Dashboard
                            </h1>
                            <p className="text-gray-600">
                                Welcome back, {admin?.name} ({admin?.role})
                            </p>
                        </div>
                    </div>

                    <div className="flex space-x-3">
                        <Link
                            to="/admin/venues"
                            className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors inline-flex items-center"
                        >
                            <Building className="h-4 w-4 mr-2" />
                            Manage Venues
                        </Link>
                        <Link
                            to="/admin/events"
                            className="bg-purple-600 text-white px-4 py-2 rounded-lg hover:bg-purple-700 transition-colors inline-flex items-center"
                        >
                            <Calendar className="h-4 w-4 mr-2" />
                            Manage Events
                        </Link>
                    </div>
                </div>
            </div>

            {/* Error Message */}
            {error && (
                <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
                    {error}
                </div>
            )}

            {/* Stats Cards */}
            <div className="grid md:grid-cols-4 gap-6">
                <div className="bg-white rounded-lg shadow-md p-6">
                    <div className="flex items-center">
                        <div className="p-3 rounded-full bg-blue-100">
                            <Calendar className="h-6 w-6 text-blue-600" />
                        </div>
                        <div className="ml-4">
                            <p className="text-sm font-medium text-gray-600">Total Events</p>
                            <p className="text-2xl font-bold text-gray-900">{stats.totalEvents}</p>
                        </div>
                    </div>
                </div>

                <div className="bg-white rounded-lg shadow-md p-6">
                    <div className="flex items-center">
                        <div className="p-3 rounded-full bg-green-100">
                            <TrendingUp className="h-6 w-6 text-green-600" />
                        </div>
                        <div className="ml-4">
                            <p className="text-sm font-medium text-gray-600">Published</p>
                            <p className="text-2xl font-bold text-gray-900">{stats.publishedEvents}</p>
                        </div>
                    </div>
                </div>

                <div className="bg-white rounded-lg shadow-md p-6">
                    <div className="flex items-center">
                        <div className="p-3 rounded-full bg-yellow-100">
                            <Settings className="h-6 w-6 text-yellow-600" />
                        </div>
                        <div className="ml-4">
                            <p className="text-sm font-medium text-gray-600">Draft Events</p>
                            <p className="text-2xl font-bold text-gray-900">{stats.draftEvents}</p>
                        </div>
                    </div>
                </div>

                <div className="bg-white rounded-lg shadow-md p-6">
                    <div className="flex items-center">
                        <div className="p-3 rounded-full bg-purple-100">
                            <Building className="h-6 w-6 text-purple-600" />
                        </div>
                        <div className="ml-4">
                            <p className="text-sm font-medium text-gray-600">Total Venues</p>
                            <p className="text-2xl font-bold text-gray-900">{stats.totalVenues}</p>
                        </div>
                    </div>
                </div>
            </div>

            {/* Quick Actions */}
            <div className="bg-white rounded-lg shadow-md p-6">
                <h3 className="text-lg font-semibold mb-4">Quick Actions</h3>
                <div className="grid md:grid-cols-3 gap-4">
                    <Link
                        to="/admin/venues"
                        className="p-4 border border-gray-200 rounded-lg hover:border-blue-300 hover:bg-blue-50 transition-colors"
                    >
                        <div className="flex items-center">
                            <Plus className="h-5 w-5 text-blue-600 mr-3" />
                            <div>
                                <p className="font-medium text-gray-900">Create Venue</p>
                                <p className="text-sm text-gray-600">Add a new venue location</p>
                            </div>
                        </div>
                    </Link>

                    <Link
                        to="/admin/events"
                        className="p-4 border border-gray-200 rounded-lg hover:border-purple-300 hover:bg-purple-50 transition-colors"
                    >
                        <div className="flex items-center">
                            <Plus className="h-5 w-5 text-purple-600 mr-3" />
                            <div>
                                <p className="font-medium text-gray-900">Create Event</p>
                                <p className="text-sm text-gray-600">Add a new event</p>
                            </div>
                        </div>
                    </Link>

                    <Link
                        to="/events"
                        className="p-4 border border-gray-200 rounded-lg hover:border-green-300 hover:bg-green-50 transition-colors"
                    >
                        <div className="flex items-center">
                            <Eye className="h-5 w-5 text-green-600 mr-3" />
                            <div>
                                <p className="font-medium text-gray-900">View Public Events</p>
                                <p className="text-sm text-gray-600">See customer view</p>
                            </div>
                        </div>
                    </Link>
                </div>
            </div>

            {/* Recent Events */}
            <div className="bg-white rounded-lg shadow-md p-6">
                <div className="flex justify-between items-center mb-4">
                    <h3 className="text-lg font-semibold">Recent Events</h3>
                    <Link
                        to="/admin/events"
                        className="text-purple-600 hover:text-purple-700 text-sm font-medium"
                    >
                        View All
                    </Link>
                </div>

                {recentEvents.length > 0 ? (
                    <div className="space-y-4">
                        {recentEvents.slice(0, 5).map((event) => (
                            <div key={event.event_id} className="flex items-center justify-between p-4 border border-gray-200 rounded-lg">
                                <div className="flex-1">
                                    <h4 className="font-medium text-gray-900">{event.name}</h4>
                                    <div className="flex items-center space-x-4 text-sm text-gray-600 mt-1">
                                        <span>{event.venue_name}</span>
                                        <span>•</span>
                                        <span>{formatDate(event.start_datetime)}</span>
                                        <span>•</span>
                                        <span>{event.available_seats} seats available</span>
                                    </div>
                                </div>

                                <div className="flex items-center space-x-3">
                                    <span className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(event.status)}`}>
                                        {event.status}
                                    </span>
                                    <span className="text-lg font-bold text-green-600">
                                        ${event.base_price}
                                    </span>
                                </div>
                            </div>
                        ))}
                    </div>
                ) : (
                    <div className="text-center py-8">
                        <Calendar className="h-12 w-12 text-gray-400 mx-auto mb-3" />
                        <p className="text-gray-600">No events created yet</p>
                        <Link
                            to="/admin/events"
                            className="mt-2 text-purple-600 hover:text-purple-700 font-medium"
                        >
                            Create your first event
                        </Link>
                    </div>
                )}
            </div>

            {/* Help Section */}
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-6">
                <h3 className="text-lg font-semibold text-blue-900 mb-2">
                    Getting Started
                </h3>
                <div className="text-blue-800 space-y-2">
                    <p>1. <strong>Create venues</strong> where your events will be held</p>
                    <p>2. <strong>Create events</strong> and assign them to venues</p>
                    <p>3. <strong>Publish events</strong> to make them available for booking</p>
                    <p>4. <strong>Monitor bookings</strong> and manage your events</p>
                </div>
            </div>
        </div>
    );
};

export default AdminDashboard;
