import React, { useState, useEffect } from 'react';
import { eventService, formatError } from '../../services/api';
import { Calendar, Plus, Edit, Trash2, Eye, Users, DollarSign } from 'lucide-react';

const AdminEvents = () => {
    const [events, setEvents] = useState([]);
    const [venues, setVenues] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [showForm, setShowForm] = useState(false);
    const [editingEvent, setEditingEvent] = useState(null);
    const [formData, setFormData] = useState({
        name: '',
        description: '',
        venue_id: '',
        event_type: 'concert',
        start_datetime: '',
        end_datetime: '',
        total_capacity: '',
        base_price: '',
        max_tickets_per_booking: '8'
    });

    useEffect(() => {
        fetchData();
    }, []);

    const fetchData = async () => {
        try {
            const [eventsResponse, venuesResponse] = await Promise.all([
                eventService.getAdminEvents(),
                eventService.getVenues()
            ]);

            setEvents(eventsResponse.data.events || []);
            setVenues(venuesResponse.data.venues || []);
        } catch (error) {
            setError(formatError(error));
        } finally {
            setLoading(false);
        }
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        setLoading(true);
        setError('');

        try {
            const eventData = {
                ...formData,
                total_capacity: parseInt(formData.total_capacity),
                base_price: parseFloat(formData.base_price),
                max_tickets_per_booking: parseInt(formData.max_tickets_per_booking),
                // Convert datetime-local format to ISO string
                start_datetime: new Date(formData.start_datetime).toISOString(),
                end_datetime: new Date(formData.end_datetime).toISOString()
            };

            if (editingEvent) {
                await eventService.updateEvent(editingEvent.event_id, eventData);
            } else {
                await eventService.createEvent(eventData);
            }

            await fetchData();
            resetForm();
        } catch (error) {
            setError(formatError(error));
        } finally {
            setLoading(false);
        }
    };

    const handlePublish = async (eventId, currentStatus) => {
        try {
            const event = events.find(e => e.event_id === eventId);
            if (!event) return;

            const newStatus = currentStatus === 'published' ? 'draft' : 'published';

            await eventService.updateEvent(eventId, {
                status: newStatus,
                version: event.version
            });

            await fetchData();
        } catch (error) {
            setError(formatError(error));
        }
    };

    const handleEdit = (event) => {
        setEditingEvent(event);
        setFormData({
            name: event.name,
            description: event.description || '',
            venue_id: event.venue_id,
            event_type: event.event_type,
            start_datetime: event.start_datetime ? event.start_datetime.slice(0, 16) : '',
            end_datetime: event.end_datetime ? event.end_datetime.slice(0, 16) : '',
            total_capacity: event.total_capacity.toString(),
            base_price: event.base_price.toString(),
            max_tickets_per_booking: event.max_tickets_per_booking.toString()
        });
        setShowForm(true);
    };

    const handleDelete = async (event) => {
        if (!window.confirm('Are you sure you want to delete this event?')) {
            return;
        }

        try {
            await eventService.deleteEvent(event.event_id, event.version);
            await fetchData();
        } catch (error) {
            setError(formatError(error));
        }
    };

    const resetForm = () => {
        setFormData({
            name: '',
            description: '',
            venue_id: '',
            event_type: 'concert',
            start_datetime: '',
            end_datetime: '',
            total_capacity: '',
            base_price: '',
            max_tickets_per_booking: '8'
        });
        setShowForm(false);
        setEditingEvent(null);
    };

    const handleChange = (e) => {
        setFormData({
            ...formData,
            [e.target.name]: e.target.value
        });
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

    if (loading && events.length === 0) {
        return (
            <div className="flex justify-center items-center min-h-[400px]">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-purple-600"></div>
            </div>
        );
    }

    return (
        <div className="max-w-7xl mx-auto space-y-6">
            {/* Header */}
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900">Event Management</h1>
                    <p className="text-gray-600">Create and manage your events</p>
                </div>

                <button
                    onClick={() => setShowForm(!showForm)}
                    className="bg-purple-600 text-white px-4 py-2 rounded-lg hover:bg-purple-700 transition-colors inline-flex items-center"
                >
                    <Plus className="h-4 w-4 mr-2" />
                    Add Event
                </button>
            </div>

            {/* Error Message */}
            {error && (
                <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
                    {error}
                </div>
            )}

            {/* Create/Edit Form */}
            {showForm && (
                <div className="bg-white rounded-lg shadow-md p-6">
                    <h3 className="text-lg font-semibold mb-4">
                        {editingEvent ? 'Edit Event' : 'Create New Event'}
                    </h3>

                    <form onSubmit={handleSubmit} className="space-y-4">
                        <div className="grid md:grid-cols-2 gap-4">
                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-1">
                                    Event Name *
                                </label>
                                <input
                                    type="text"
                                    name="name"
                                    value={formData.name}
                                    onChange={handleChange}
                                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500"
                                    required
                                />
                            </div>

                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-1">
                                    Event Type *
                                </label>
                                <select
                                    name="event_type"
                                    value={formData.event_type}
                                    onChange={handleChange}
                                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500"
                                    required
                                >
                                    <option value="concert">Concert</option>
                                    <option value="sports">Sports</option>
                                    <option value="theater">Theater</option>
                                    <option value="comedy">Comedy</option>
                                    <option value="festival">Festival</option>
                                    <option value="conference">Conference</option>
                                    <option value="other">Other</option>
                                </select>
                            </div>
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">
                                Description
                            </label>
                            <textarea
                                name="description"
                                value={formData.description}
                                onChange={handleChange}
                                rows="3"
                                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500"
                                placeholder="Event description..."
                            />
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">
                                Venue *
                            </label>
                            <select
                                name="venue_id"
                                value={formData.venue_id}
                                onChange={handleChange}
                                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500"
                                required
                            >
                                <option value="">Select a venue</option>
                                {venues.map(venue => (
                                    <option key={venue.venue_id} value={venue.venue_id}>
                                        {venue.name} - {venue.city} (Capacity: {venue.capacity})
                                    </option>
                                ))}
                            </select>
                        </div>

                        <div className="grid md:grid-cols-2 gap-4">
                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-1">
                                    Start Date & Time *
                                </label>
                                <input
                                    type="datetime-local"
                                    name="start_datetime"
                                    value={formData.start_datetime}
                                    onChange={handleChange}
                                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500"
                                    required
                                />
                            </div>

                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-1">
                                    End Date & Time *
                                </label>
                                <input
                                    type="datetime-local"
                                    name="end_datetime"
                                    value={formData.end_datetime}
                                    onChange={handleChange}
                                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500"
                                    required
                                />
                            </div>
                        </div>

                        <div className="grid md:grid-cols-3 gap-4">
                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-1">
                                    Total Capacity *
                                </label>
                                <input
                                    type="number"
                                    name="total_capacity"
                                    value={formData.total_capacity}
                                    onChange={handleChange}
                                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500"
                                    required
                                    min="1"
                                />
                            </div>

                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-1">
                                    Base Price ($) *
                                </label>
                                <input
                                    type="number"
                                    name="base_price"
                                    value={formData.base_price}
                                    onChange={handleChange}
                                    step="0.01"
                                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500"
                                    required
                                    min="0"
                                />
                            </div>

                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-1">
                                    Max Tickets per Booking *
                                </label>
                                <input
                                    type="number"
                                    name="max_tickets_per_booking"
                                    value={formData.max_tickets_per_booking}
                                    onChange={handleChange}
                                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500"
                                    required
                                    min="1"
                                    max="50"
                                />
                            </div>
                        </div>

                        <div className="flex space-x-3">
                            <button
                                type="submit"
                                disabled={loading}
                                className="bg-purple-600 text-white px-4 py-2 rounded-md hover:bg-purple-700 transition-colors disabled:opacity-50"
                            >
                                {loading ? 'Saving...' : editingEvent ? 'Update Event' : 'Create Event'}
                            </button>
                            <button
                                type="button"
                                onClick={resetForm}
                                className="border border-gray-300 text-gray-700 px-4 py-2 rounded-md hover:bg-gray-50 transition-colors"
                            >
                                Cancel
                            </button>
                        </div>
                    </form>
                </div>
            )}

            {/* Events List */}
            <div className="bg-white rounded-lg shadow-md">
                <div className="p-6 border-b border-gray-200">
                    <h3 className="text-lg font-semibold">
                        All Events ({events.length})
                    </h3>
                </div>

                {events.length > 0 ? (
                    <div className="divide-y divide-gray-200">
                        {events.map((event) => (
                            <div key={event.event_id} className="p-6">
                                <div className="flex justify-between items-start">
                                    <div className="flex-1">
                                        <div className="flex items-center mb-2">
                                            <Calendar className="h-5 w-5 text-purple-600 mr-2" />
                                            <h4 className="text-lg font-semibold text-gray-900">
                                                {event.name}
                                            </h4>
                                            <span className={`ml-3 px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(event.status)}`}>
                                                {event.status}
                                            </span>
                                        </div>

                                        <div className="grid md:grid-cols-2 gap-4 text-sm text-gray-600 mb-3">
                                            <div>
                                                <p><strong>Venue:</strong> {event.venue_name}</p>
                                                <p><strong>Type:</strong> {event.event_type}</p>
                                                <p><strong>Start:</strong> {formatDate(event.start_datetime)}</p>
                                            </div>
                                            <div>
                                                <div className="flex items-center">
                                                    <Users className="h-4 w-4 mr-1" />
                                                    <span>{event.available_seats} / {event.total_capacity} available</span>
                                                </div>
                                                <div className="flex items-center">
                                                    <DollarSign className="h-4 w-4 mr-1" />
                                                    <span>${event.base_price}</span>
                                                </div>
                                                <p><strong>Max per booking:</strong> {event.max_tickets_per_booking}</p>
                                            </div>
                                        </div>

                                        {event.description && (
                                            <p className="text-sm text-gray-600 mb-3">
                                                {event.description.length > 150
                                                    ? `${event.description.substring(0, 150)}...`
                                                    : event.description
                                                }
                                            </p>
                                        )}
                                    </div>

                                    <div className="flex space-x-2 ml-4">
                                        <button
                                            onClick={() => handlePublish(event.event_id, event.status)}
                                            className={`px-3 py-1 rounded text-sm font-medium ${event.status === 'published'
                                                ? 'bg-yellow-100 text-yellow-800 hover:bg-yellow-200'
                                                : 'bg-green-100 text-green-800 hover:bg-green-200'
                                                }`}
                                        >
                                            {event.status === 'published' ? 'Unpublish' : 'Publish'}
                                        </button>

                                        <button
                                            onClick={() => window.open(`/events/${event.event_id}`, '_blank')}
                                            className="text-blue-600 hover:text-blue-700 p-1"
                                            title="View public page"
                                        >
                                            <Eye className="h-4 w-4" />
                                        </button>

                                        <button
                                            onClick={() => handleEdit(event)}
                                            className="text-gray-600 hover:text-gray-700 p-1"
                                            title="Edit event"
                                        >
                                            <Edit className="h-4 w-4" />
                                        </button>

                                        <button
                                            onClick={() => handleDelete(event)}
                                            className="text-red-600 hover:text-red-700 p-1"
                                            title="Delete event"
                                        >
                                            <Trash2 className="h-4 w-4" />
                                        </button>
                                    </div>
                                </div>
                            </div>
                        ))}
                    </div>
                ) : (
                    <div className="text-center py-12">
                        <Calendar className="h-16 w-16 text-gray-400 mx-auto mb-4" />
                        <h3 className="text-lg font-semibold text-gray-900 mb-2">
                            No events created yet
                        </h3>
                        <p className="text-gray-600 mb-6">
                            {venues.length === 0
                                ? 'Create a venue first, then add your events'
                                : 'Create your first event to start accepting bookings'
                            }
                        </p>
                        {venues.length === 0 ? (
                            <a
                                href="/admin/venues"
                                className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 transition-colors"
                            >
                                Create Venue First
                            </a>
                        ) : (
                            <button
                                onClick={() => setShowForm(true)}
                                className="bg-purple-600 text-white px-6 py-2 rounded-lg hover:bg-purple-700 transition-colors"
                            >
                                Create Event
                            </button>
                        )}
                    </div>
                )}
            </div>
        </div>
    );
};

export default AdminEvents;
