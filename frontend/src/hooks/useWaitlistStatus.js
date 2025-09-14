import { useState, useEffect, useCallback, useRef } from 'react';
import { bookingService, formatError } from '../services/api';

const useWaitlistStatus = (eventId, isAuthenticated, options = {}) => {
    const {
        pollInterval = 5000, // 5 seconds
        onStatusChange = () => {},
        onOffer = () => {},
        onOfferExpired = () => {}
    } = options;

    const [status, setStatus] = useState(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const intervalRef = useRef(null);
    const previousStatusRef = useRef(null);

    const fetchStatus = useCallback(async () => {
        if (!isAuthenticated || !eventId) {
            setStatus(null);
            setLoading(false);
            return;
        }

        try {
            const response = await bookingService.getWaitlistPosition({
                event_id: eventId
            });
            
            const newStatus = response.data;
            const previousStatus = previousStatusRef.current;
            
            // Check for status changes
            if (previousStatus && previousStatus.status !== newStatus.status) {
                onStatusChange(newStatus, previousStatus);
                
                // Check if user received an offer
                if (newStatus.status === 'offered' && previousStatus.status === 'waiting') {
                    onOffer(newStatus);
                }
                
                // Check if offer expired
                if (newStatus.status === 'waiting' && previousStatus.status === 'offered') {
                    onOfferExpired(newStatus);
                }
            }
            
            setStatus(newStatus);
            previousStatusRef.current = newStatus;
            setError('');
            
        } catch (error) {
            const errorMessage = formatError(error);
            
            // If user is not on waitlist, that's not an error
            if (errorMessage.includes('Not in waitlist') || errorMessage.includes('not found')) {
                setStatus(null);
                setError('');
            } else {
                setError(errorMessage);
            }
        } finally {
            setLoading(false);
        }
    }, [eventId, isAuthenticated, onStatusChange, onOffer, onOfferExpired]);

    // Start polling
    const startPolling = useCallback(() => {
        if (intervalRef.current) return; // Already polling
        
        fetchStatus(); // Immediate fetch
        intervalRef.current = setInterval(fetchStatus, pollInterval);
    }, [fetchStatus, pollInterval]);

    // Stop polling
    const stopPolling = useCallback(() => {
        if (intervalRef.current) {
            clearInterval(intervalRef.current);
            intervalRef.current = null;
        }
    }, []);

    // Manual refresh
    const refresh = useCallback(() => {
        setLoading(true);
        fetchStatus();
    }, [fetchStatus]);

    // Start/stop polling based on authentication and component mount
    useEffect(() => {
        if (isAuthenticated && eventId) {
            startPolling();
        } else {
            stopPolling();
            setStatus(null);
            setLoading(false);
        }

        return () => {
            stopPolling();
        };
    }, [isAuthenticated, eventId, startPolling, stopPolling]);

    // Cleanup on unmount
    useEffect(() => {
        return () => {
            stopPolling();
        };
    }, [stopPolling]);

    return {
        status,
        loading,
        error,
        refresh,
        startPolling,
        stopPolling,
        isPolling: intervalRef.current !== null
    };
};

export default useWaitlistStatus;