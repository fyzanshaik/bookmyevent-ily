import React, { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { searchService, formatError } from '../services/api';
import { Search, Filter, MapPin, Calendar, DollarSign, Users, X } from 'lucide-react';

const Events = () => {
    const [events, setEvents] = useState([]);
    const [filters, setFilters] = useState({
        cities: [],
        event_types: [],
        price_range: { min: 0, max: 1000 }
    });
    const [searchParams, setSearchParams] = useState({
        q: '',
        city: '',
        type: '',
        min_price: '',
        max_price: '',
        page: 1,
        limit: 20
    });
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [showFilters, setShowFilters] = useState(false);
    const [totalResults, setTotalResults] = useState(0);

    // Fetch filter options
    useEffect(() => {
        const fetchFilters = async () => {
            try {
                const response = await searchService.getFilters();
                setFilters(response.data);
            } catch (error) {
                console.error('Failed to fetch filters:', formatError(error));
            }
        };
        fetchFilters();
    }, []);

    // Search events
    const searchEvents = useCallback(async (params = searchParams) => {
        setLoading(true);
        setError('');

        try {
            // Clean up empty parameters
            const cleanParams = Object.entries(params)
                .filter(([, value]) => value !== '' && value !== null && value !== undefined)
                .reduce((acc, [key, value]) => ({ ...acc, [key]: value }), {});

            const response = await searchService.search(cleanParams);
            setEvents(response.data.results || []);
            setTotalResults(response.data.total || 0);
        } catch (error) {
            setError(formatError(error));
            setEvents([]);
        } finally {
            setLoading(false);
        }
    }, [searchParams]);

    // Initial load
    useEffect(() => {
        searchEvents();
    }, [searchEvents]);

    const handleSearch = (e) => {
        e.preventDefault();
        setSearchParams({ ...searchParams, page: 1 });
        searchEvents({ ...searchParams, page: 1 });
    };

    const handleFilterChange = (key, value) => {
        const newParams = { ...searchParams, [key]: value, page: 1 };
        setSearchParams(newParams);
        searchEvents(newParams);
    };

    const clearFilters = () => {
        const clearedParams = {
            q: searchParams.q,
            city: '',
            type: '',
            min_price: '',
            max_price: '',
            page: 1,
            limit: 20
        };
        setSearchParams(clearedParams);
        searchEvents(clearedParams);
    };

    const formatDate = (dateString) => {
        return new Date(dateString).toLocaleDateString('en-US', {
            weekday: 'short',
            year: 'numeric',
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    };

    return (
        <div className="space-y-6">
            {/* Header */}
            <div className="text-center">
                <h1 className="text-3xl font-bold text-gray-900 mb-2">
                    Discover Events
                </h1>
                <p className="text-gray-600">
                    Find and book tickets for amazing events near you
                </p>
            </div>

            {/* Search Bar */}
            <form onSubmit={handleSearch} className="bg-white p-6 rounded-lg shadow-md">
                <div className="flex flex-col md:flex-row gap-4">
                    <div className="flex-1 relative">
                        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-gray-400" />
                        <input
                            type="text"
                            placeholder="Search events, venues, or artists..."
                            value={searchParams.q}
                            onChange={(e) => setSearchParams({ ...searchParams, q: e.target.value })}
                            className="w-full pl-10 pr-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                        />
                    </div>
                    <button
                        type="submit"
                        className="bg-blue-600 text-white px-8 py-3 rounded-lg hover:bg-blue-700 transition-colors"
                    >
                        Search
                    </button>
                    <button
                        type="button"
                        onClick={() => setShowFilters(!showFilters)}
                        className="border border-gray-300 text-gray-700 px-6 py-3 rounded-lg hover:bg-gray-50 transition-colors inline-flex items-center"
                    >
                        <Filter className="h-5 w-5 mr-2" />
                        Filters
                    </button>
                </div>
            </form>

            {/* Filters Panel */}
            {showFilters && (
                <div className="bg-white p-6 rounded-lg shadow-md">
                    <div className="flex justify-between items-center mb-4">
                        <h3 className="text-lg font-semibold">Filters</h3>
                        <button
                            onClick={clearFilters}
                            className="text-blue-600 hover:text-blue-700 text-sm"
                        >
                            Clear All
                        </button>
                    </div>

                    <div className="grid md:grid-cols-4 gap-4">
                        {/* City Filter */}
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-2">
                                City
                            </label>
                            <select
                                value={searchParams.city}
                                onChange={(e) => handleFilterChange('city', e.target.value)}
                                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                            >
                                <option value="">All Cities</option>
                                {filters.cities.map(city => (
                                    <option key={city} value={city}>{city}</option>
                                ))}
                            </select>
                        </div>

                        {/* Event Type Filter */}
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-2">
                                Event Type
                            </label>
                            <select
                                value={searchParams.type}
                                onChange={(e) => handleFilterChange('type', e.target.value)}
                                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                            >
                                <option value="">All Types</option>
                                {filters.event_types.map(type => (
                                    <option key={type} value={type}>{type}</option>
                                ))}
                            </select>
                        </div>

                        {/* Min Price Filter */}
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-2">
                                Min Price ($)
                            </label>
                            <input
                                type="number"
                                placeholder="0"
                                value={searchParams.min_price}
                                onChange={(e) => handleFilterChange('min_price', e.target.value)}
                                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                            />
                        </div>

                        {/* Max Price Filter */}
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-2">
                                Max Price ($)
                            </label>
                            <input
                                type="number"
                                placeholder="1000"
                                value={searchParams.max_price}
                                onChange={(e) => handleFilterChange('max_price', e.target.value)}
                                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                            />
                        </div>
                    </div>
                </div>
            )}

            {/* Active Filters */}
            {(searchParams.city || searchParams.type || searchParams.min_price || searchParams.max_price) && (
                <div className="flex flex-wrap gap-2">
                    {searchParams.city && (
                        <span className="inline-flex items-center bg-blue-100 text-blue-800 px-3 py-1 rounded-full text-sm">
                            City: {searchParams.city}
                            <X
                                className="h-4 w-4 ml-2 cursor-pointer"
                                onClick={() => handleFilterChange('city', '')}
                            />
                        </span>
                    )}
                    {searchParams.type && (
                        <span className="inline-flex items-center bg-green-100 text-green-800 px-3 py-1 rounded-full text-sm">
                            Type: {searchParams.type}
                            <X
                                className="h-4 w-4 ml-2 cursor-pointer"
                                onClick={() => handleFilterChange('type', '')}
                            />
                        </span>
                    )}
                    {searchParams.min_price && (
                        <span className="inline-flex items-center bg-yellow-100 text-yellow-800 px-3 py-1 rounded-full text-sm">
                            Min: ${searchParams.min_price}
                            <X
                                className="h-4 w-4 ml-2 cursor-pointer"
                                onClick={() => handleFilterChange('min_price', '')}
                            />
                        </span>
                    )}
                    {searchParams.max_price && (
                        <span className="inline-flex items-center bg-red-100 text-red-800 px-3 py-1 rounded-full text-sm">
                            Max: ${searchParams.max_price}
                            <X
                                className="h-4 w-4 ml-2 cursor-pointer"
                                onClick={() => handleFilterChange('max_price', '')}
                            />
                        </span>
                    )}
                </div>
            )}

            {/* Results Count */}
            {!loading && (
                <div className="text-sm text-gray-600">
                    {totalResults > 0
                        ? `Showing ${events.length} of ${totalResults} events`
                        : 'No events found'
                    }
                </div>
            )}

            {/* Error Message */}
            {error && (
                <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
                    {error}
                </div>
            )}

            {/* Events Grid */}
            {loading ? (
                <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
                    {[...Array(6)].map((_, i) => (
                        <div key={i} className="bg-white rounded-lg shadow-md p-6 animate-pulse">
                            <div className="h-6 bg-gray-200 rounded mb-3"></div>
                            <div className="h-4 bg-gray-200 rounded mb-2"></div>
                            <div className="h-4 bg-gray-200 rounded mb-4"></div>
                            <div className="h-10 bg-gray-200 rounded"></div>
                        </div>
                    ))}
                </div>
            ) : events.length > 0 ? (
                <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
                    {events.map((event) => (
                        <div key={event.event_id} className="bg-white rounded-lg shadow-md overflow-hidden hover:shadow-lg transition-shadow">
                            <div className="p-6">
                                <h3 className="text-xl font-semibold text-gray-900 mb-3 line-clamp-2">
                                    {event.name}
                                </h3>

                                <div className="space-y-2 mb-4">
                                    <div className="flex items-center text-gray-600 text-sm">
                                        <MapPin className="h-4 w-4 mr-2" />
                                        <span>{event.venue_name}, {event.venue_city}</span>
                                    </div>

                                    <div className="flex items-center text-gray-600 text-sm">
                                        <Calendar className="h-4 w-4 mr-2" />
                                        <span>{formatDate(event.start_datetime)}</span>
                                    </div>

                                    <div className="flex items-center text-gray-600 text-sm">
                                        <Users className="h-4 w-4 mr-2" />
                                        <span>{event.available_seats} seats available</span>
                                    </div>
                                </div>

                                <div className="flex justify-between items-center mb-4">
                                    <div className="flex items-center">
                                        <DollarSign className="h-5 w-5 text-green-600" />
                                        <span className="text-2xl font-bold text-green-600">
                                            {event.base_price}
                                        </span>
                                    </div>

                                    <span className="bg-blue-100 text-blue-800 px-2 py-1 rounded-full text-xs font-medium">
                                        {event.event_type}
                                    </span>
                                </div>

                                <Link
                                    to={`/events/${event.event_id}`}
                                    className="w-full bg-blue-600 text-white py-3 px-4 rounded-lg hover:bg-blue-700 transition-colors inline-flex items-center justify-center font-semibold"
                                >
                                    View Details & Book
                                </Link>
                            </div>
                        </div>
                    ))}
                </div>
            ) : (
                <div className="text-center py-12">
                    <Search className="h-16 w-16 text-gray-400 mx-auto mb-4" />
                    <h3 className="text-xl font-semibold text-gray-900 mb-2">
                        No events found
                    </h3>
                    <p className="text-gray-600 mb-6">
                        Try adjusting your search criteria or removing some filters
                    </p>
                    <button
                        onClick={clearFilters}
                        className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 transition-colors"
                    >
                        Clear Filters
                    </button>
                </div>
            )}
        </div>
    );
};

export default Events;
