import React, { useState, useEffect, useCallback } from 'react';
import { useParams, useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { bookingService, eventService, formatError } from '../services/api';
import {
    CreditCard, Clock, CheckCircle, AlertCircle,
    ArrowLeft, Ticket, DollarSign
} from 'lucide-react';

const BookingFlow = () => {
    const { eventId } = useParams();
    const location = useLocation();
    const navigate = useNavigate();
    const { user } = useAuth();

    const [step, setStep] = useState(1);
    const [event, setEvent] = useState(null);
    const [quantity, setQuantity] = useState(location.state?.quantity || 1);
    const [reservation, setReservation] = useState(null);
    const [booking, setBooking] = useState(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [timeLeft, setTimeLeft] = useState(0);

    // Payment form state
    const [paymentData, setPaymentData] = useState({
        payment_method: 'credit_card',
        payment_token: 'mock-payment-token'
    });

    useEffect(() => {
        const fetchEvent = async () => {
            try {
                const response = await eventService.getEvent(eventId);
                setEvent(response.data);
            } catch (error) {
                setError(formatError(error));
            }
        };
        fetchEvent();
    }, [eventId]);

    const handleReservationExpiry = useCallback(async () => {
        if (!reservation) return;

        try {
            await bookingService.expireReservation(reservation.reservation_id);
            console.log('Reservation manually expired and seats returned');
            setReservation(null); // Clear reservation to stop timer
            setError('Reservation has been expired. Please start over.');
        } catch (error) {
            console.error('Manual expiry failed:', formatError(error));
            console.log('Note: Expired reservation cleanup will be handled by background process');
            setReservation(null); // Clear reservation anyway
            setError('Reservation expired. Please start over.');
        }
    }, [reservation]);

    // Countdown timer for reservation
    useEffect(() => {
        if (reservation && reservation.expires_at) {
            // Check immediately if already expired (handles page refresh)
            const checkExpiry = () => {
                const now = new Date().getTime();
                const expiry = new Date(reservation.expires_at).getTime();
                const remaining = expiry - now;

                if (remaining <= 0) {
                    setTimeLeft(0);
                    setError('Your reservation has expired. Please start over.');
                    handleReservationExpiry();
                    setReservation(null); // Clear reservation to stop timer
                    return false; // Expired
                } else {
                    setTimeLeft(Math.floor(remaining / 1000));
                    return true; // Still valid
                }
            };

            // Check immediately on load
            const isValid = checkExpiry();

            if (isValid) {
                // Only start interval if not expired
                const interval = setInterval(checkExpiry, 1000);
                return () => clearInterval(interval);
            }
        }
    }, [reservation, handleReservationExpiry]);

    const formatTime = (seconds) => {
        const minutes = Math.floor(seconds / 60);
        const remainingSeconds = seconds % 60;
        return `${minutes}:${remainingSeconds.toString().padStart(2, '0')}`;
    };

    const handleReserve = async () => {
        setLoading(true);
        setError('');

        try {
            const idempotencyKey = `${user.user_id}-booking-${eventId}-${Date.now()}`;

            const response = await bookingService.reserve({
                event_id: eventId,
                quantity: quantity,
                idempotency_key: idempotencyKey
            });

            setReservation(response.data);
            setStep(2);
        } catch (error) {
            setError(formatError(error));
        } finally {
            setLoading(false);
        }
    };

    const handlePayment = async () => {
        if (!reservation) return;

        setLoading(true);
        setError('');

        try {
            const response = await bookingService.confirm({
                reservation_id: reservation.reservation_id,
                payment_token: paymentData.payment_token,
                payment_method: paymentData.payment_method
            });

            setBooking(response.data);
            setStep(3);
        } catch (error) {
            setError(formatError(error));
        } finally {
            setLoading(false);
        }
    };

    if (!event) {
        return (
            <div className="flex justify-center items-center min-h-[400px]">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
            </div>
        );
    }

    const totalAmount = (event.base_price * quantity).toFixed(2);

    return (
        <div className="max-w-2xl mx-auto space-y-6">
            {/* Back Button */}
            <button
                onClick={() => navigate(-1)}
                className="inline-flex items-center text-gray-600 hover:text-blue-600 transition-colors"
            >
                <ArrowLeft className="h-5 w-5 mr-2" />
                Back to Event
            </button>

            {/* Progress Bar */}
            <div className="bg-white rounded-lg shadow-md p-6">
                <div className="flex items-center justify-between mb-4">
                    <div className={`flex items-center ${step >= 1 ? 'text-blue-600' : 'text-gray-400'}`}>
                        <div className={`w-8 h-8 rounded-full flex items-center justify-center ${step >= 1 ? 'bg-blue-600 text-white' : 'bg-gray-300'}`}>
                            1
                        </div>
                        <span className="ml-2 font-medium">Reserve</span>
                    </div>

                    <div className={`flex-1 h-1 mx-4 ${step >= 2 ? 'bg-blue-600' : 'bg-gray-300'}`}></div>

                    <div className={`flex items-center ${step >= 2 ? 'text-blue-600' : 'text-gray-400'}`}>
                        <div className={`w-8 h-8 rounded-full flex items-center justify-center ${step >= 2 ? 'bg-blue-600 text-white' : 'bg-gray-300'}`}>
                            2
                        </div>
                        <span className="ml-2 font-medium">Payment</span>
                    </div>

                    <div className={`flex-1 h-1 mx-4 ${step >= 3 ? 'bg-blue-600' : 'bg-gray-300'}`}></div>

                    <div className={`flex items-center ${step >= 3 ? 'text-blue-600' : 'text-gray-400'}`}>
                        <div className={`w-8 h-8 rounded-full flex items-center justify-center ${step >= 3 ? 'bg-blue-600 text-white' : 'bg-gray-300'}`}>
                            3
                        </div>
                        <span className="ml-2 font-medium">Confirm</span>
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

            {/* Step 1: Reserve */}
            {step === 1 && (
                <div className="bg-white rounded-lg shadow-md p-6">
                    <h2 className="text-2xl font-bold mb-6">Reserve Your Tickets</h2>

                    <div className="space-y-4 mb-6">
                        <div className="flex justify-between items-center p-4 bg-gray-50 rounded">
                            <span className="font-medium">Event:</span>
                            <span>{event.name}</span>
                        </div>

                        <div className="flex justify-between items-center p-4 bg-gray-50 rounded">
                            <span className="font-medium">Quantity:</span>
                            <div className="flex items-center space-x-2">
                                <button
                                    onClick={() => setQuantity(Math.max(1, quantity - 1))}
                                    className="w-8 h-8 rounded-full bg-gray-300 hover:bg-gray-400 flex items-center justify-center"
                                >
                                    -
                                </button>
                                <span className="w-8 text-center">{quantity}</span>
                                <button
                                    onClick={() => setQuantity(Math.min(10, quantity + 1))}
                                    className="w-8 h-8 rounded-full bg-gray-300 hover:bg-gray-400 flex items-center justify-center"
                                >
                                    +
                                </button>
                            </div>
                        </div>

                        <div className="flex justify-between items-center p-4 bg-gray-50 rounded">
                            <span className="font-medium">Price per ticket:</span>
                            <span>${event.base_price}</span>
                        </div>

                        <div className="flex justify-between items-center p-4 bg-blue-50 rounded border border-blue-200">
                            <span className="font-bold text-lg">Total:</span>
                            <span className="font-bold text-lg text-blue-600">${totalAmount}</span>
                        </div>
                    </div>

                    <div className="bg-yellow-50 border border-yellow-200 rounded p-4 mb-6">
                        <div className="flex items-center">
                            <Clock className="h-5 w-5 text-yellow-600 mr-2" />
                            <p className="text-yellow-800 text-sm">
                                <strong>Important:</strong> You'll have 5 minutes to complete your payment after reservation.
                            </p>
                        </div>
                    </div>

                    <button
                        onClick={handleReserve}
                        disabled={loading}
                        className="w-full bg-blue-600 text-white py-3 px-4 rounded-lg hover:bg-blue-700 transition-colors font-semibold disabled:opacity-50"
                    >
                        {loading ? 'Reserving...' : 'Reserve Tickets'}
                    </button>
                </div>
            )}

            {/* Step 2: Payment */}
            {step === 2 && reservation && (
                <div className="bg-white rounded-lg shadow-md p-6">
                    <div className="flex justify-between items-center mb-6">
                        <h2 className="text-2xl font-bold">Complete Payment</h2>
                        {timeLeft > 0 && (
                            <div className="flex items-center gap-4">
                                <div className="flex items-center text-red-600">
                                    <Clock className="h-5 w-5 mr-2" />
                                    <span className="font-mono text-lg">{formatTime(timeLeft)}</span>
                                </div>
                                <button
                                    onClick={handleReservationExpiry}
                                    className="px-3 py-1 text-xs bg-red-100 text-red-600 rounded hover:bg-red-200 transition-colors"
                                    title="Force expiry for testing"
                                >
                                    Force Expiry
                                </button>
                            </div>
                        )}
                    </div>

                    <div className="bg-green-50 border border-green-200 rounded p-4 mb-6">
                        <div className="flex items-center">
                            <CheckCircle className="h-5 w-5 text-green-600 mr-2" />
                            <div>
                                <p className="text-green-800 font-medium">Tickets Reserved!</p>
                                <p className="text-green-700 text-sm">
                                    Reservation ID: {reservation.booking_reference}
                                </p>
                            </div>
                        </div>
                    </div>

                    <div className="space-y-4 mb-6">
                        <div className="flex justify-between items-center p-4 bg-gray-50 rounded">
                            <span className="font-medium">Total Amount:</span>
                            <span className="font-bold text-lg">${reservation.total_amount}</span>
                        </div>
                    </div>

                    {/* Mock Payment Form */}
                    <div className="space-y-4 mb-6">
                        <h3 className="text-lg font-semibold">Payment Method</h3>

                        <div className="space-y-3">
                            <label className="flex items-center">
                                <input
                                    type="radio"
                                    name="payment_method"
                                    value="credit_card"
                                    checked={paymentData.payment_method === 'credit_card'}
                                    onChange={(e) => setPaymentData({ ...paymentData, payment_method: e.target.value })}
                                    className="mr-3"
                                />
                                <CreditCard className="h-5 w-5 mr-2" />
                                Credit Card
                            </label>

                            <label className="flex items-center">
                                <input
                                    type="radio"
                                    name="payment_method"
                                    value="debit_card"
                                    checked={paymentData.payment_method === 'debit_card'}
                                    onChange={(e) => setPaymentData({ ...paymentData, payment_method: e.target.value })}
                                    className="mr-3"
                                />
                                <CreditCard className="h-5 w-5 mr-2" />
                                Debit Card
                            </label>

                            <label className="flex items-center">
                                <input
                                    type="radio"
                                    name="payment_method"
                                    value="paypal"
                                    checked={paymentData.payment_method === 'paypal'}
                                    onChange={(e) => setPaymentData({ ...paymentData, payment_method: e.target.value })}
                                    className="mr-3"
                                />
                                <DollarSign className="h-5 w-5 mr-2" />
                                PayPal
                            </label>
                        </div>

                        <div className="bg-blue-50 border border-blue-200 rounded p-4">
                            <p className="text-blue-800 text-sm">
                                <strong>Demo Mode:</strong> This is a simulated payment. No actual charges will be made.
                            </p>
                        </div>
                    </div>

                    <button
                        onClick={handlePayment}
                        disabled={loading || timeLeft <= 0}
                        className="w-full bg-green-600 text-white py-3 px-4 rounded-lg hover:bg-green-700 transition-colors font-semibold disabled:opacity-50"
                    >
                        {loading ? 'Processing Payment...' : `Pay $${reservation.total_amount}`}
                    </button>
                </div>
            )}

            {/* Step 3: Confirmation */}
            {step === 3 && booking && (
                <div className="bg-white rounded-lg shadow-md p-6">
                    <div className="text-center mb-6">
                        <CheckCircle className="h-16 w-16 text-green-600 mx-auto mb-4" />
                        <h2 className="text-2xl font-bold text-green-600 mb-2">
                            Booking Confirmed!
                        </h2>
                        <p className="text-gray-600">
                            Your tickets have been successfully booked
                        </p>
                    </div>

                    <div className="space-y-4 mb-6">
                        <div className="flex justify-between items-center p-4 bg-gray-50 rounded">
                            <span className="font-medium">Booking Reference:</span>
                            <span className="font-mono font-bold">{booking.booking_reference}</span>
                        </div>

                        <div className="flex justify-between items-center p-4 bg-gray-50 rounded">
                            <span className="font-medium">Event:</span>
                            <span>{event.name}</span>
                        </div>

                        <div className="flex justify-between items-center p-4 bg-gray-50 rounded">
                            <span className="font-medium">Quantity:</span>
                            <span>{quantity} ticket{quantity > 1 ? 's' : ''}</span>
                        </div>

                        <div className="flex justify-between items-center p-4 bg-gray-50 rounded">
                            <span className="font-medium">Total Paid:</span>
                            <span className="font-bold">${booking.payment?.amount || booking.total_amount}</span>
                        </div>

                        {booking.ticket_url && (
                            <div className="p-4 bg-blue-50 rounded border border-blue-200">
                                <div className="flex items-center">
                                    <Ticket className="h-5 w-5 text-blue-600 mr-2" />
                                    <div>
                                        <p className="text-blue-800 font-medium">Digital Tickets</p>
                                        <a
                                            href={booking.ticket_url}
                                            target="_blank"
                                            rel="noopener noreferrer"
                                            className="text-blue-600 hover:text-blue-700 text-sm underline"
                                        >
                                            View/Download Tickets
                                        </a>
                                    </div>
                                </div>
                            </div>
                        )}
                    </div>

                    <div className="flex space-x-4">
                        <button
                            onClick={() => navigate('/dashboard')}
                            className="flex-1 bg-blue-600 text-white py-3 px-4 rounded-lg hover:bg-blue-700 transition-colors font-semibold"
                        >
                            View My Bookings
                        </button>
                        <button
                            onClick={() => navigate('/events')}
                            className="flex-1 border border-gray-300 text-gray-700 py-3 px-4 rounded-lg hover:bg-gray-50 transition-colors font-semibold"
                        >
                            Browse More Events
                        </button>
                    </div>
                </div>
            )}
        </div>
    );
};

export default BookingFlow;
