import React, { useState, useEffect, useCallback } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { eventService, bookingService, formatError } from '../services/api';
import {
    MapPin, Calendar, Users, DollarSign, Clock,
    Ticket, AlertCircle, CheckCircle, ArrowLeft
} from 'lucide-react';

const EventDetails = () => {
    const { eventId } = useParams();
    const { isAuthenticated } = useAuth();
    const navigate = useNavigate();

    const [event, setEvent] = useState(null);
    const [availability, setAvailability] = useState(null);
    const [loading, setLoading] = useState(true);
    const [availabilityLoading, setAvailabilityLoading] = useState(false);
    const [error, setError] = useState('');
    const [quantity, setQuantity] = useState(1);

    useEffect(() => {
        const fetchEventDetails = async () => {
            try {
                const response = await eventService.getEvent(eventId);
                setEvent(response.data);
            } catch (error) {
                setError(formatError(error));
            } finally {
                setLoading(false);
            }
        };

        fetchEventDetails();
    }, [eventId]);

    const checkAvailability = useCallback(async (requestedQuantity = quantity) => {
        if (!event) return;

        setAvailabilityLoading(true);
        try {
            const response = await bookingService.checkAvailability({
                event_id: eventId,
                quantity: requestedQuantity
            });
            setAvailability(response.data);
        } catch (error) {
            console.error('Availability check failed:', formatError(error));
            setAvailability(null);
        } finally {
            setAvailabilityLoading(false);
        }
    }, [event, eventId, quantity]);

    useEffect(() => {
        if (event) {
            checkAvailability();
        }
    }, [event, quantity, checkAvailability]);

    const handleBookNow = () => {
        if (!isAuthenticated) {
            navigate('/login', { state: { from: { pathname: `/book/${eventId}` } } });
            return;
        }
        navigate(`/book/${eventId}`, { state: { quantity } });
    };

    const formatDate = (dateString) => {
        return new Date(dateString).toLocaleDateString('en-US', {
            weekday: 'long',
            year: 'numeric',
            month: 'long',
            day: 'numeric'
        });
    };

    const formatTime = (dateString) => {
        return new Date(dateString).toLocaleTimeString('en-US', {
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

    if (error || !event) {
        return (
            <div className="text-center py-12">
                <AlertCircle className="h-16 w-16 text-red-500 mx-auto mb-4" />
                <h2 className="text-2xl font-bold text-gray-900 mb-2">Event Not Found</h2>
                <p className="text-gray-600 mb-6">{error || 'The event you are looking for does not exist.'}</p>
                <Link to="/events" className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 transition-colors">
                    Browse Events
                </Link>
            </div>
        );
    }

    const totalPrice = (availability?.base_price || event.base_price) * quantity;
    const isAvailable = availability?.available && availability.available_seats >= quantity;

    return (
        <div className="max-w-4xl mx-auto space-y-8">
            {/* Back Button */}
            <button
                onClick={() => navigate(-1)}
                className="inline-flex items-center text-gray-600 hover:text-blue-600 transition-colors"
            >
                <ArrowLeft className="h-5 w-5 mr-2" />
                Back to Events
            </button>

            {/* Event Header */}
            <div className="bg-white rounded-lg shadow-md p-8">
                <div className="grid md:grid-cols-3 gap-8">
                    {/* Event Info */}
                    <div className="md:col-span-2 space-y-4">
                        <h1 className="text-3xl font-bold text-gray-900">
                            {event.name}
                        </h1>

                        {event.description && (
                            <p className="text-gray-600 text-lg">
                                {event.description}
                            </p>
                        )}

                        <div className="space-y-3">
                            <div className="flex items-center text-gray-700">
                                <MapPin className="h-5 w-5 mr-3 text-blue-600" />
                                <div>
                                    <p className="font-semibold">{event.venue_name}</p>
                                    <p className="text-sm text-gray-600">
                                        {event.venue_address}, {event.venue_city}
                                    </p>
                                </div>
                            </div>

                            <div className="flex items-center text-gray-700">
                                <Calendar className="h-5 w-5 mr-3 text-blue-600" />
                                <div>
                                    <p className="font-semibold">{formatDate(event.start_datetime)}</p>
                                    <p className="text-sm text-gray-600">
                                        {formatTime(event.start_datetime)} - {formatTime(event.end_datetime)}
                                    </p>
                                </div>
                            </div>

                            <div className="flex items-center text-gray-700">
                                <Users className="h-5 w-5 mr-3 text-blue-600" />
                                <div>
                                    <p className="font-semibold">
                                        {availability?.available_seats || event.available_seats} seats available
                                    </p>
                                    <p className="text-sm text-gray-600">
                                        Total capacity: {event.total_capacity}
                                    </p>
                                </div>
                            </div>

                            <div className="flex items-center text-gray-700">
                                <DollarSign className="h-5 w-5 mr-3 text-blue-600" />
                                <div>
                                    <p className="font-semibold text-2xl text-green-600">
                                        ${availability?.base_price || event.base_price}
                                    </p>
                                    <p className="text-sm text-gray-600">Starting price per ticket</p>
                                </div>
                            </div>

                            <div className="flex items-center text-gray-700">
                                <Ticket className="h-5 w-5 mr-3 text-blue-600" />
                                <div>
                                    <p className="font-semibold">
                                        Max {availability?.max_per_booking || event.max_tickets_per_booking} tickets per booking
                                    </p>
                                    <p className="text-sm text-gray-600">Booking limit per user</p>
                                </div>
                            </div>
                        </div>

                        <div className="pt-4">
                            <span className="inline-block bg-blue-100 text-blue-800 px-3 py-1 rounded-full text-sm font-medium">
                                {event.event_type}
                            </span>
                            <span className="inline-block bg-green-100 text-green-800 px-3 py-1 rounded-full text-sm font-medium ml-2">
                                {event.status}
                            </span>
                        </div>
                    </div>

                    {/* Booking Panel */}
                    <div className="bg-gray-50 rounded-lg p-6">
                        <h3 className="text-xl font-semibold mb-4">Book Tickets</h3>

                        {/* Quantity Selector */}
                        <div className="mb-4">
                            <label className="block text-sm font-medium text-gray-700 mb-2">
                                Number of Tickets
                            </label>
                            <select
                                value={quantity}
                                onChange={(e) => setQuantity(parseInt(e.target.value))}
                                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                            >
                                {[...Array(Math.min(10, availability?.max_per_booking || 10))].map((_, i) => (
                                    <option key={i + 1} value={i + 1}>
                                        {i + 1} ticket{i > 0 ? 's' : ''}
                                    </option>
                                ))}
                            </select>
                        </div>

                        {/* Availability Status */}
                        <div className="mb-4">
                            {availabilityLoading ? (
                                <div className="flex items-center text-gray-600">
                                    <Clock className="h-4 w-4 mr-2 animate-spin" />
                                    Checking availability...
                                </div>
                            ) : availability ? (
                                isAvailable ? (
                                    <div className="flex items-center text-green-600">
                                        <CheckCircle className="h-4 w-4 mr-2" />
                                        {quantity} ticket{quantity > 1 ? 's' : ''} available
                                    </div>
                                ) : (
                                    <div className="flex items-center text-red-600">
                                        <AlertCircle className="h-4 w-4 mr-2" />
                                        Not enough seats available
                                    </div>
                                )
                            ) : (
                                <div className="flex items-center text-red-600">
                                    <AlertCircle className="h-4 w-4 mr-2" />
                                    Unable to check availability
                                </div>
                            )}
                        </div>

                        {/* Total Price */}
                        <div className="mb-6 p-3 bg-white rounded border">
                            <div className="flex justify-between items-center">
                                <span className="text-gray-600">
                                    {quantity} × ${availability?.base_price || event.base_price}
                                </span>
                                <span className="text-xl font-bold text-green-600">
                                    ${totalPrice.toFixed(2)}
                                </span>
                            </div>
                        </div>

                        {/* Book Button */}
                        {event.status === 'published' ? (
                            isAvailable ? (
                                <button
                                    onClick={handleBookNow}
                                    disabled={availabilityLoading}
                                    className="w-full bg-blue-600 text-white py-3 px-4 rounded-lg hover:bg-blue-700 transition-colors font-semibold disabled:opacity-50"
                                >
                                    {isAuthenticated ? 'Book Now' : 'Login to Book'}
                                </button>
                            ) : (
                                <div className="space-y-2">
                                    <button
                                        disabled
                                        className="w-full bg-gray-400 text-white py-3 px-4 rounded-lg font-semibold cursor-not-allowed"
                                    >
                                        Not Available
                                    </button>
                                    {isAuthenticated && (
                                        <Link
                                            to={`/waitlist/${eventId}`}
                                            className="w-full bg-yellow-600 text-white py-2 px-4 rounded-lg hover:bg-yellow-700 transition-colors font-semibold text-center block"
                                        >
                                            Join Waitlist
                                        </Link>
                                    )}
                                </div>
                            )
                        ) : (
                            <button
                                disabled
                                className="w-full bg-gray-400 text-white py-3 px-4 rounded-lg font-semibold cursor-not-allowed"
                            >
                                Event Not Available
                            </button>
                        )}

                        {!isAuthenticated && (
                            <p className="text-sm text-gray-600 mt-3 text-center">
                                <Link to="/login" className="text-blue-600 hover:text-blue-700">
                                    Login
                                </Link>{' '}
                                or{' '}
                                <Link to="/register" className="text-blue-600 hover:text-blue-700">
                                    Register
                                </Link>{' '}
                                to book tickets
                            </p>
                        )}
                    </div>
                </div>
            </div>

            {/* Additional Info */}
            <div className="bg-white rounded-lg shadow-md p-8">
                <h3 className="text-xl font-semibold mb-4">Event Information</h3>

                <div className="grid md:grid-cols-2 gap-6">
                    <div>
                        <h4 className="font-semibold text-gray-900 mb-2">Venue Details</h4>
                        <p className="text-gray-600">{event.venue_name}</p>
                        {event.venue_address && (
                            <p className="text-gray-600">{event.venue_address}</p>
                        )}
                        <p className="text-gray-600">{event.venue_city}, {event.venue_state}</p>
                        <p className="text-gray-600">{event.venue_country}</p>
                    </div>

                    <div>
                        <h4 className="font-semibold text-gray-900 mb-2">Booking Policy</h4>
                        <ul className="text-gray-600 space-y-1 text-sm">
                            <li>• Maximum {availability?.max_per_booking || event.max_tickets_per_booking} tickets per booking</li>
                            <li>• 5-minute reservation window</li>
                            <li>• Secure payment processing</li>
                            <li>• Instant ticket delivery</li>
                        </ul>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default EventDetails;
