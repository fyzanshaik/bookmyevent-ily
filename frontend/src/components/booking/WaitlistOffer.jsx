import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { CheckCircle, Clock, AlertTriangle, Ticket } from 'lucide-react';

const WaitlistOffer = ({ 
    eventId, 
    eventName, 
    quantity, 
    offeredAt,
    expiresAt, 
    onExpired = () => {},
    onAccept = () => {},
    onDecline = () => {} 
}) => {
    const navigate = useNavigate();
    const [timeLeft, setTimeLeft] = useState(0);
    const [totalDuration, setTotalDuration] = useState(0);
    const [isExpired, setIsExpired] = useState(false);

    useEffect(() => {
        if (!expiresAt || !offeredAt) return;

        const calculateTime = () => {
            const now = new Date().getTime();
            const expiry = new Date(expiresAt).getTime();
            const offer = new Date(offeredAt).getTime();

            if (isNaN(expiry) || isNaN(offer)) {
                return;
            }

            const total = Math.floor((expiry - offer) / 1000);
            const remaining = Math.floor((expiry - now) / 1000);

            setTotalDuration(total > 0 ? total : 0);

            if (remaining <= 0) {
                setTimeLeft(0);
                if (!isExpired) {
                    setIsExpired(true);
                    onExpired();
                }
                return;
            }
            setTimeLeft(remaining);
        };

        calculateTime();
        const interval = setInterval(calculateTime, 1000);
        return () => clearInterval(interval);
    }, [expiresAt, offeredAt, onExpired, isExpired]);


    const formatTime = (seconds) => {
        const mins = Math.floor(seconds / 60);
        const secs = seconds % 60;
        return `${mins}:${secs.toString().padStart(2, '0')}`;
    };

    const handleAcceptOffer = () => {
        onAccept();
        navigate(`/book/${eventId}?quantity=${quantity}&waitlist_offer=true`);
    };

    const handleDeclineOffer = () => {
        onDecline();
    };

    const getProgressPercentage = () => {
        if (totalDuration === 0) return 0;
        const progress = ((totalDuration - timeLeft) / totalDuration) * 100;
        return Math.min(Math.max(progress, 0), 100);
    };

    const getProgressColor = () => {
        const percentage = getProgressPercentage();
        if (percentage > 75) return 'bg-red-500';
        if (percentage > 50) return 'bg-yellow-500';
        return 'bg-green-500';
    };

    if (isExpired) {
        return (
            <div className="bg-red-50 border border-red-200 rounded-lg p-6">
                <div className="flex items-center justify-center mb-4">
                    <AlertTriangle className="h-12 w-12 text-red-600 mr-3" />
                    <div>
                        <h3 className="text-xl font-semibold text-red-600">Offer Expired</h3>
                        <p className="text-red-700">Your booking window has closed</p>
                    </div>
                </div>
                
                <p className="text-red-700 text-center mb-4">
                    The booking window has expired. You've been returned to the waitlist.
                </p>
                
                <button
                    onClick={() => window.location.reload()}
                    className="w-full bg-red-600 text-white py-2 px-4 rounded-lg hover:bg-red-700 transition-colors font-medium"
                >
                    Return to Waitlist
                </button>
            </div>
        );
    }

    return (
        <div className="bg-green-50 border border-green-200 rounded-lg p-6">
            <div className="flex items-center justify-center mb-6">
                <CheckCircle className="h-12 w-12 text-green-600 mr-3" />
                <div>
                    <h3 className="text-xl font-semibold text-green-600">🎉 Tickets Available!</h3>
                    <p className="text-green-700">Your turn to book has arrived</p>
                </div>
            </div>

            <div className="bg-white rounded-lg p-4 mb-6">
                <div className="flex items-center mb-2">
                    <Ticket className="h-5 w-5 text-blue-600 mr-2" />
                    <span className="font-semibold">{eventName}</span>
                </div>
                <p className="text-gray-600">
                    {quantity} ticket{quantity > 1 ? 's' : ''} available for you
                </p>
            </div>

            <div className="mb-6">
                <div className="flex items-center justify-center mb-3">
                    <Clock className="h-6 w-6 text-orange-600 mr-2" />
                    <div className="text-center">
                        <div className="text-3xl font-bold text-orange-600">
                            {formatTime(timeLeft)}
                        </div>
                        <div className="text-sm text-orange-700">
                            Time remaining to complete booking
                        </div>
                    </div>
                </div>

                <div className="w-full bg-gray-200 rounded-full h-3">
                    <div
                        className={`h-3 rounded-full transition-all duration-1000 ${getProgressColor()}`}
                        style={{ width: `${getProgressPercentage()}%` }}
                    ></div>
                </div>
            </div>

            <div className="bg-orange-50 border border-orange-200 rounded p-4 mb-6">
                <div className="flex items-start">
                    <AlertTriangle className="h-5 w-5 text-orange-600 mr-2 mt-0.5" />
                    <div>
                        <p className="text-orange-800 font-medium">Important!</p>
                        <ul className="text-orange-700 text-sm mt-1 space-y-1">
                            <li>• You have a limited time to complete your booking</li>
                            <li>• If time expires, the tickets go to the next person in line</li>
                            <li>• Click "Book Now" to reserve your tickets immediately</li>
                        </ul>
                    </div>
                </div>
            </div>

            <div className="flex space-x-3">
                <button
                    onClick={handleAcceptOffer}
                    className="flex-1 bg-green-600 text-white py-3 px-4 rounded-lg hover:bg-green-700 transition-colors font-semibold"
                >
                    Book Now ({quantity} ticket{quantity > 1 ? 's' : ''})
                </button>
                <button
                    onClick={handleDeclineOffer}
                    className="border border-gray-300 text-gray-700 py-3 px-4 rounded-lg hover:bg-gray-50 transition-colors font-medium"
                >
                    Decline
                </button>
            </div>

            <p className="text-xs text-gray-600 text-center mt-4">
                This offer is exclusively for you and cannot be shared
            </p>
        </div>
    );
};

export default WaitlistOffer;