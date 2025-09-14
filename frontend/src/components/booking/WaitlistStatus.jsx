import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { bookingService, formatError } from '../../services/api';
import { Clock, Users, AlertCircle, CheckCircle, Info } from 'lucide-react';

const WaitlistStatus = ({ eventId, eventName, onLeave }) => {
    const [status, setStatus] = useState(null);
    const [loading, setLoading] = useState(true);
    const [leaving, setLeaving] = useState(false);
    const [error, setError] = useState('');

    useEffect(() => {
        const fetchStatus = async () => {
            try {
                const response = await bookingService.getWaitlistPosition({
                    event_id: eventId
                });
                setStatus(response.data);
            } catch (error) {
                // User not on waitlist
                setStatus(null);
            } finally {
                setLoading(false);
            }
        };

        fetchStatus();
    }, [eventId]);

    const handleLeave = async () => {
        setLeaving(true);
        setError('');

        try {
            await bookingService.leaveWaitlist({
                event_id: eventId
            });
            
            setStatus(null);
            if (onLeave) {
                onLeave(eventId);
            }
        } catch (error) {
            setError(formatError(error));
        } finally {
            setLeaving(false);
        }
    };

    if (loading) {
        return (
            <div className="bg-white p-4 rounded-lg shadow-md">
                <div className="animate-pulse flex space-x-4">
                    <div className="rounded-full bg-gray-200 h-10 w-10"></div>
                    <div className="flex-1 space-y-2 py-1">
                        <div className="h-4 bg-gray-200 rounded w-3/4"></div>
                        <div className="h-4 bg-gray-200 rounded w-1/2"></div>
                    </div>
                </div>
            </div>
        );
    }

    if (!status) {
        return null; // User is not on waitlist
    }

    const getStatusColor = (status) => {
        switch (status) {
            case 'waiting':
                return 'text-blue-600';
            case 'offered':
                return 'text-green-600';
            case 'expired':
                return 'text-red-600';
            default:
                return 'text-gray-600';
        }
    };

    const getStatusIcon = (status) => {
        switch (status) {
            case 'waiting':
                return <Clock className="h-5 w-5 text-blue-600" />;
            case 'offered':
                return <CheckCircle className="h-5 w-5 text-green-600" />;
            case 'expired':
                return <AlertCircle className="h-5 w-5 text-red-600" />;
            default:
                return <Info className="h-5 w-5 text-gray-600" />;
        }
    };

    return (
        <div className="bg-white p-6 rounded-lg shadow-md">
            <div className="flex items-center justify-between mb-4">
                <div className="flex items-center">
                    {getStatusIcon(status.status)}
                    <h3 className="ml-2 text-lg font-semibold">Waitlist Status</h3>
                </div>
                <span className={`px-2 py-1 text-xs font-medium rounded-full capitalize ${getStatusColor(status.status)} bg-gray-100`}>
                    {status.status}
                </span>
            </div>

            <div className="space-y-3 mb-4">
                <div>
                    <p className="text-sm text-gray-600">Event</p>
                    <p className="font-medium">{eventName}</p>
                </div>

                <div className="grid grid-cols-2 gap-4">
                    <div>
                        <p className="text-sm text-gray-600">Position</p>
                        <p className="text-xl font-bold text-blue-600">#{status.position}</p>
                    </div>
                    <div>
                        <p className="text-sm text-gray-600">Estimated Wait</p>
                        <p className="font-medium">{status.estimated_wait || 'TBD'}</p>
                    </div>
                </div>
            </div>

            {status.status === 'offered' && (
                <div className="bg-green-50 border border-green-200 rounded p-3 mb-4">
                    <div className="flex items-start">
                        <CheckCircle className="h-5 w-5 text-green-600 mr-2 mt-0.5" />
                        <div>
                            <p className="text-green-800 font-medium">Tickets Available!</p>
                            <p className="text-green-700 text-sm">
                                You have been offered tickets. Complete your booking within the time limit.
                            </p>
                        </div>
                    </div>
                </div>
            )}

            {status.status === 'waiting' && (
                <div className="bg-blue-50 border border-blue-200 rounded p-3 mb-4">
                    <div className="flex items-start">
                        <Clock className="h-5 w-5 text-blue-600 mr-2 mt-0.5" />
                        <div>
                            <p className="text-blue-800 font-medium">You're in the queue</p>
                            <p className="text-blue-700 text-sm">
                                We'll notify you when tickets become available.
                            </p>
                        </div>
                    </div>
                </div>
            )}

            {error && (
                <div className="bg-red-100 border border-red-400 text-red-700 px-3 py-2 rounded mb-4 text-sm flex items-center">
                    <AlertCircle className="h-4 w-4 mr-2" />
                    {error}
                </div>
            )}

            <div className="flex space-x-3">
                <Link
                    to={`/events/${eventId}`}
                    className="flex-1 text-center bg-blue-600 text-white py-2 px-4 rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium"
                >
                    View Event
                </Link>
                <button
                    onClick={handleLeave}
                    disabled={leaving}
                    className="flex-1 border border-red-300 text-red-600 py-2 px-4 rounded-lg hover:bg-red-50 transition-colors text-sm font-medium disabled:opacity-50"
                >
                    {leaving ? 'Leaving...' : 'Leave Waitlist'}
                </button>
            </div>
        </div>
    );
};

export default WaitlistStatus;