import React, { useState, useEffect } from 'react';
import { eventService, formatError } from '../../services/api';
import { Building, Plus, Edit, Trash2, MapPin, Users } from 'lucide-react';

const AdminVenues = () => {
    const [venues, setVenues] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [showForm, setShowForm] = useState(false);
    const [editingVenue, setEditingVenue] = useState(null);
    const [formData, setFormData] = useState({
        name: '',
        address: '',
        city: '',
        state: '',
        country: '',
        postal_code: '',
        capacity: ''
    });

    useEffect(() => {
        fetchVenues();
    }, []);

    const fetchVenues = async () => {
        try {
            const response = await eventService.getVenues();
            setVenues(response.data.venues || []);
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
            const venueData = {
                ...formData,
                capacity: parseInt(formData.capacity)
            };

            if (editingVenue) {
                await eventService.updateVenue(editingVenue.venue_id, venueData);
            } else {
                await eventService.createVenue(venueData);
            }

            await fetchVenues();
            resetForm();
        } catch (error) {
            setError(formatError(error));
        } finally {
            setLoading(false);
        }
    };

    const handleEdit = (venue) => {
        setEditingVenue(venue);
        setFormData({
            name: venue.name,
            address: venue.address || '',
            city: venue.city,
            state: venue.state || '',
            country: venue.country,
            postal_code: venue.postal_code || '',
            capacity: venue.capacity.toString()
        });
        setShowForm(true);
    };

    const handleDelete = async (venueId) => {
        if (!window.confirm('Are you sure you want to delete this venue?')) {
            return;
        }

        try {
            await eventService.deleteVenue(venueId);
            await fetchVenues();
        } catch (error) {
            setError(formatError(error));
        }
    };

    const resetForm = () => {
        setFormData({
            name: '',
            address: '',
            city: '',
            state: '',
            country: '',
            postal_code: '',
            capacity: ''
        });
        setShowForm(false);
        setEditingVenue(null);
    };

    const handleChange = (e) => {
        setFormData({
            ...formData,
            [e.target.name]: e.target.value
        });
    };

    if (loading && venues.length === 0) {
        return (
            <div className="flex justify-center items-center min-h-[400px]">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-purple-600"></div>
            </div>
        );
    }

    return (
        <div className="max-w-6xl mx-auto space-y-6">
            {/* Header */}
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900">Venue Management</h1>
                    <p className="text-gray-600">Create and manage event venues</p>
                </div>

                <button
                    onClick={() => setShowForm(!showForm)}
                    className="bg-purple-600 text-white px-4 py-2 rounded-lg hover:bg-purple-700 transition-colors inline-flex items-center"
                >
                    <Plus className="h-4 w-4 mr-2" />
                    Add Venue
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
                        {editingVenue ? 'Edit Venue' : 'Create New Venue'}
                    </h3>

                    <form onSubmit={handleSubmit} className="grid md:grid-cols-2 gap-4">
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">
                                Venue Name *
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
                                Capacity *
                            </label>
                            <input
                                type="number"
                                name="capacity"
                                value={formData.capacity}
                                onChange={handleChange}
                                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500"
                                required
                                min="1"
                            />
                        </div>

                        <div className="md:col-span-2">
                            <label className="block text-sm font-medium text-gray-700 mb-1">
                                Address
                            </label>
                            <input
                                type="text"
                                name="address"
                                value={formData.address}
                                onChange={handleChange}
                                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500"
                                placeholder="123 Main Street"
                            />
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">
                                City *
                            </label>
                            <input
                                type="text"
                                name="city"
                                value={formData.city}
                                onChange={handleChange}
                                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500"
                                required
                            />
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">
                                State/Province
                            </label>
                            <input
                                type="text"
                                name="state"
                                value={formData.state}
                                onChange={handleChange}
                                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500"
                            />
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">
                                Country *
                            </label>
                            <input
                                type="text"
                                name="country"
                                value={formData.country}
                                onChange={handleChange}
                                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500"
                                required
                            />
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">
                                Postal Code
                            </label>
                            <input
                                type="text"
                                name="postal_code"
                                value={formData.postal_code}
                                onChange={handleChange}
                                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500"
                            />
                        </div>

                        <div className="md:col-span-2 flex space-x-3">
                            <button
                                type="submit"
                                disabled={loading}
                                className="bg-purple-600 text-white px-4 py-2 rounded-md hover:bg-purple-700 transition-colors disabled:opacity-50"
                            >
                                {loading ? 'Saving...' : editingVenue ? 'Update Venue' : 'Create Venue'}
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

            {/* Venues List */}
            <div className="bg-white rounded-lg shadow-md">
                <div className="p-6 border-b border-gray-200">
                    <h3 className="text-lg font-semibold">
                        All Venues ({venues.length})
                    </h3>
                </div>

                {venues.length > 0 ? (
                    <div className="divide-y divide-gray-200">
                        {venues.map((venue) => (
                            <div key={venue.venue_id} className="p-6">
                                <div className="flex justify-between items-start">
                                    <div className="flex-1">
                                        <div className="flex items-center mb-2">
                                            <Building className="h-5 w-5 text-purple-600 mr-2" />
                                            <h4 className="text-lg font-semibold text-gray-900">
                                                {venue.name}
                                            </h4>
                                        </div>

                                        <div className="space-y-1 text-sm text-gray-600">
                                            {venue.address && (
                                                <div className="flex items-center">
                                                    <MapPin className="h-4 w-4 mr-2" />
                                                    <span>{venue.address}</span>
                                                </div>
                                            )}

                                            <div className="flex items-center">
                                                <MapPin className="h-4 w-4 mr-2" />
                                                <span>
                                                    {venue.city}
                                                    {venue.state && `, ${venue.state}`}
                                                    {venue.country && `, ${venue.country}`}
                                                    {venue.postal_code && ` ${venue.postal_code}`}
                                                </span>
                                            </div>

                                            <div className="flex items-center">
                                                <Users className="h-4 w-4 mr-2" />
                                                <span>Capacity: {venue.capacity.toLocaleString()}</span>
                                            </div>
                                        </div>
                                    </div>

                                    <div className="flex space-x-2 ml-4">
                                        <button
                                            onClick={() => handleEdit(venue)}
                                            className="text-blue-600 hover:text-blue-700 p-1"
                                        >
                                            <Edit className="h-4 w-4" />
                                        </button>
                                        <button
                                            onClick={() => handleDelete(venue.venue_id)}
                                            className="text-red-600 hover:text-red-700 p-1"
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
                        <Building className="h-16 w-16 text-gray-400 mx-auto mb-4" />
                        <h3 className="text-lg font-semibold text-gray-900 mb-2">
                            No venues created yet
                        </h3>
                        <p className="text-gray-600 mb-6">
                            Create your first venue to start hosting events
                        </p>
                        <button
                            onClick={() => setShowForm(true)}
                            className="bg-purple-600 text-white px-6 py-2 rounded-lg hover:bg-purple-700 transition-colors"
                        >
                            Create Venue
                        </button>
                    </div>
                )}
            </div>
        </div>
    );
};

export default AdminVenues;
