import React, { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import useWaitlistStatus from '../hooks/useWaitlistStatus';
import WaitlistOffer from '../components/booking/WaitlistOffer';
import { eventService, bookingService, formatError } from '../services/api';
import {
    Clock, Users, AlertCircle, CheckCircle, 
    ArrowLeft, Ticket, Info
} from 'lucide-react';

const WaitlistPage = () => {
    const { eventId } = useParams();
    const navigate = useNavigate();
    const { user, isAuthenticated } = useAuth();

    const [event, setEvent] = useState(null);
    const [eventLoading, setEventLoading] = useState(true);
    const [joining, setJoining] = useState(false);
    const [leaving, setLeaving] = useState(false);
    const [error, setError] = useState('');
    const [quantity, setQuantity] = useState(1);

    // Waitlist status hook with polling and status change detection
    const { 
        status: waitlistStatus, 
        loading: statusLoading, 
        error: statusError,
        refresh: refreshStatus 
    } = useWaitlistStatus(eventId, isAuthenticated, {
        onOffer: useCallback((newStatus) => {
            console.log('🎉 Received waitlist offer!', newStatus);
        }, []),
        onOfferExpired: useCallback((newStatus) => {
            console.log('⏰ Waitlist offer expired, back to waiting', newStatus);
            setError('Your booking window expired. You\'re back in the waitlist.');
        }, []),
        onStatusChange: useCallback((newStatus, oldStatus) => {
            console.log('Waitlist status changed:', oldStatus?.status, '→', newStatus?.status);
        }, [])
    });

    // Fetch event details
    useEffect(() => {
        const fetchEvent = async () => {
            try {
                const response = await eventService.getEvent(eventId);
                setEvent(response.data);
            } catch (error) {
                setError(formatError(error));
            } finally {
                setEventLoading(false);
            }
        };
        fetchEvent();
    }, [eventId]);

    const handleJoinWaitlist = async () => {
        setJoining(true);
        setError('');

        try {
            await bookingService.joinWaitlist({
                event_id: eventId,
                quantity: quantity
            });
            
            // Refresh status immediately after joining
            refreshStatus();
        } catch (error) {
            setError(formatError(error));
        } finally {
            setJoining(false);
        }
    };

    const handleLeaveWaitlist = async () => {
        setLeaving(true);
        setError('');

        try {
            await bookingService.leaveWaitlist({
                event_id: eventId,
                user_id: user.user_id
            });
            
            // Refresh status immediately after leaving
            refreshStatus();
        } catch (error) {
            setError(formatError(error));
        } finally {
            setLeaving(false);
        }
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

    if (eventLoading) {
        return (
            <div className="flex justify-center items-center min-h-[400px]">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
            </div>
        );
    }

    if (error && !event) {
        return (
            <div className="text-center py-12">
                <AlertCircle className="h-16 w-16 text-red-500 mx-auto mb-4" />
                <h2 className="text-2xl font-bold text-gray-900 mb-2">Event Not Found</h2>
                <p className="text-gray-600 mb-6">{error}</p>
                <Link to="/events" className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 transition-colors">
                    Browse Events
                </Link>
            </div>
        );
    }

    if (!event) return null;

    return (
        <div className="max-w-2xl mx-auto space-y-6">
            {/* Back Button */}
            <button
                onClick={() => navigate(`/events/${eventId}`)}
                className="inline-flex items-center text-gray-600 hover:text-blue-600 transition-colors"
            >
                <ArrowLeft className="h-5 w-5 mr-2" />
                Back to Event
            </button>

            {/* Event Info */}
            <div className="bg-white rounded-lg shadow-md p-6">
                <h1 className="text-2xl font-bold text-gray-900 mb-4">{event.name}</h1>
                
                <div className="space-y-3 mb-4">
                    <div className="flex items-center text-gray-700">
                        <Ticket className="h-5 w-5 mr-3 text-blue-600" />
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
                            <p className="font-semibold text-red-600">Sold Out</p>
                            <p className="text-sm text-gray-600">
                                {event.available_seats} of {event.total_capacity} seats available
                            </p>
                        </div>
                    </div>
                </div>

                <div className="bg-yellow-50 border border-yellow-200 rounded p-4">
                    <div className="flex items-start">
                        <Info className="h-5 w-5 text-yellow-600 mr-2 mt-0.5" />
                        <div>
                            <p className="text-yellow-800 font-medium">Event is sold out</p>
                            <p className="text-yellow-700 text-sm mt-1">
                                Join the waitlist to be notified when tickets become available due to cancellations.
                            </p>
                        </div>
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

            {/* Authentication Check */}
            {!isAuthenticated ? (
                <div className="bg-white rounded-lg shadow-md p-6 text-center">
                    <h2 className="text-xl font-semibold mb-4">Join Waitlist</h2>
                    <p className="text-gray-600 mb-6">
                        You need to be logged in to join the waitlist for this event.
                    </p>
                    <div className="flex gap-4 justify-center">
                        <Link
                            to="/login"
                            state={{ from: { pathname: `/waitlist/${eventId}` } }}
                            className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 transition-colors"
                        >
                            Login
                        </Link>
                        <Link
                            to="/register"
                            className="border border-gray-300 text-gray-700 px-6 py-2 rounded-lg hover:bg-gray-50 transition-colors"
                        >
                            Register
                        </Link>
                    </div>
                </div>
            ) : waitlistStatus ? (
                /* Handle different waitlist statuses */
                waitlistStatus.status === 'offered' ? (
                    /* Show Offer Component */
                    <WaitlistOffer
                        eventId={eventId}
                        eventName={event.name}
                        quantity={waitlistStatus.quantity_requested}
                        offeredAt={waitlistStatus.offered_at}
                        expiresAt={waitlistStatus.expires_at}
                        onExpired={() => {
                            setError('Your booking window expired. Please refresh to check your waitlist status.');
                        }}
                        onAccept={() => {
                            console.log('User accepted waitlist offer');
                        }}
                        onDecline={handleLeaveWaitlist}
                    />
                ) : (
                    /* Regular Waitlist Status */
                    <div className="bg-white rounded-lg shadow-md p-6">
                        <div className="flex items-center justify-center mb-6">
                            <CheckCircle className="h-12 w-12 text-green-600 mr-3" />
                            <div>
                                <h2 className="text-xl font-semibold text-green-600">You're on the waitlist!</h2>
                                <p className="text-gray-600">We'll notify you when tickets become available</p>
                            </div>
                        </div>

                        <div className="space-y-4 mb-6">
                            <div className="flex justify-between items-center p-4 bg-gray-50 rounded">
                                <span className="font-medium">Your Position:</span>
                                <span className="text-2xl font-bold text-blue-600">#{waitlistStatus.position}</span>
                            </div>

                            <div className="flex justify-between items-center p-4 bg-gray-50 rounded">
                                <span className="font-medium">Estimated Wait:</span>
                                <span className="font-semibold">{waitlistStatus.estimated_wait || 'Check back soon'}</span>
                            </div>

                            <div className="flex justify-between items-center p-4 bg-gray-50 rounded">
                                <span className="font-medium">Status:</span>
                                <span className={`capitalize font-semibold ${
                                    waitlistStatus.status === 'waiting' ? 'text-blue-600' : 
                                    waitlistStatus.status === 'offered' ? 'text-green-600' : 
                                    'text-gray-600'
                                }`}>
                                    {waitlistStatus.status}
                                </span>
                            </div>
                        </div>

                        {/* Show polling indicator when checking status */}
                        {statusLoading && (
                            <div className="bg-blue-50 border border-blue-200 rounded p-4 mb-6">
                                <div className="flex items-center">
                                    <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600 mr-3"></div>
                                    <p className="text-blue-800 text-sm">Checking for updates...</p>
                                </div>
                            </div>
                        )}

                        <div className="bg-blue-50 border border-blue-200 rounded p-4 mb-6">
                            <div className="flex items-start">
                                <Clock className="h-5 w-5 text-blue-600 mr-2 mt-0.5" />
                                <div>
                                    <p className="text-blue-800 font-medium">What happens next?</p>
                                    <p className="text-blue-700 text-sm mt-1">
                                        When tickets become available, you'll have 2 minutes to complete your booking before the offer expires.
                                    </p>
                                </div>
                            </div>
                        </div>

                        <button
                            onClick={handleLeaveWaitlist}
                            disabled={leaving}
                            className="w-full border border-red-300 text-red-600 py-3 px-4 rounded-lg hover:bg-red-50 transition-colors font-semibold disabled:opacity-50"
                        >
                            {leaving ? 'Leaving...' : 'Leave Waitlist'}
                        </button>
                    </div>
                )
            ) : (
                /* Join Waitlist Form */
                <div className="bg-white rounded-lg shadow-md p-6">
                    <h2 className="text-xl font-semibold mb-6">Join Waitlist</h2>

                    <div className="mb-6">
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                            Number of Tickets
                        </label>
                        <select
                            value={quantity}
                            onChange={(e) => setQuantity(parseInt(e.target.value))}
                            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                        >
                            {[...Array(Math.min(10, event.max_tickets_per_booking))].map((_, i) => (
                                <option key={i + 1} value={i + 1}>
                                    {i + 1} ticket{i > 0 ? 's' : ''}
                                </option>
                            ))}
                        </select>
                    </div>

                    <div className="bg-blue-50 border border-blue-200 rounded p-4 mb-6">
                        <div className="flex items-start">
                            <Info className="h-5 w-5 text-blue-600 mr-2 mt-0.5" />
                            <div>
                                <p className="text-blue-800 font-medium">Waitlist Terms</p>
                                <ul className="text-blue-700 text-sm mt-1 space-y-1">
                                    <li>• You'll be notified when tickets become available</li>
                                    <li>• You'll have 2 minutes to complete your purchase</li>
                                    <li>• Position is based on first-come, first-served</li>
                                    <li>• You can leave the waitlist at any time</li>
                                </ul>
                            </div>
                        </div>
                    </div>

                    <button
                        onClick={handleJoinWaitlist}
                        disabled={joining}
                        className="w-full bg-blue-600 text-white py-3 px-4 rounded-lg hover:bg-blue-700 transition-colors font-semibold disabled:opacity-50"
                    >
                        {joining ? 'Joining...' : `Join Waitlist for ${quantity} ticket${quantity > 1 ? 's' : ''}`}
                    </button>
                </div>
            )}
        </div>
    );
};

export default WaitlistPage;