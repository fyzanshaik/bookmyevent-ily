import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { searchService, formatError } from '../services/api';
import { Calendar, Search, Users, Star, ArrowRight, Ticket, Shield } from 'lucide-react';

const Home = () => {
    const { isAuthenticated, isAdminAuthenticated } = useAuth();
    const [trendingEvents, setTrendingEvents] = useState([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchTrendingEvents = async () => {
            try {
                const response = await searchService.getTrending({ limit: 6 });
                setTrendingEvents(response.data.events || []);
            } catch (error) {
                console.error('Failed to fetch trending events:', formatError(error));
            } finally {
                setLoading(false);
            }
        };

        fetchTrendingEvents();
    }, []);

    return (
        <div className="space-y-16">
            {/* Hero Section */}
            <section className="text-center py-16 bg-gradient-to-r from-blue-600 to-purple-600 text-white rounded-lg">
                <div className="max-w-4xl mx-auto px-6">
                    <Calendar className="h-16 w-16 mx-auto mb-6" />
                    <h1 className="text-5xl font-bold mb-6">
                        Book Your Perfect Event
                    </h1>
                    <p className="text-xl mb-8 opacity-90">
                        Discover amazing events, book tickets instantly, and create unforgettable memories
                    </p>
                    <div className="flex flex-col sm:flex-row gap-4 justify-center">
                        <Link
                            to="/events"
                            className="bg-white text-blue-600 px-8 py-3 rounded-lg font-semibold hover:bg-gray-100 transition-colors inline-flex items-center justify-center"
                        >
                            <Search className="h-5 w-5 mr-2" />
                            Explore Events
                        </Link>
                        {!isAuthenticated && !isAdminAuthenticated && (
                            <Link
                                to="/register"
                                className="border-2 border-white text-white px-8 py-3 rounded-lg font-semibold hover:bg-white hover:text-blue-600 transition-colors inline-flex items-center justify-center"
                            >
                                <Users className="h-5 w-5 mr-2" />
                                Get Started
                            </Link>
                        )}
                    </div>
                </div>
            </section>

            {/* Features Section */}
            <section className="py-16">
                <div className="text-center mb-12">
                    <h2 className="text-3xl font-bold text-gray-900 mb-4">
                        Why Choose BookMyEvent?
                    </h2>
                    <p className="text-gray-600 max-w-2xl mx-auto">
                        Experience the most reliable and user-friendly event booking platform
                    </p>
                </div>

                <div className="grid md:grid-cols-3 gap-8">
                    <div className="text-center p-6 bg-white rounded-lg shadow-md">
                        <Ticket className="h-12 w-12 text-blue-600 mx-auto mb-4" />
                        <h3 className="text-xl font-semibold mb-3">Zero Overselling</h3>
                        <p className="text-gray-600">
                            Advanced concurrency control ensures you get the tickets you reserve
                        </p>
                    </div>

                    <div className="text-center p-6 bg-white rounded-lg shadow-md">
                        <Search className="h-12 w-12 text-green-600 mx-auto mb-4" />
                        <h3 className="text-xl font-semibold mb-3">Lightning Fast Search</h3>
                        <p className="text-gray-600">
                            Find events in under 25ms with our advanced search technology
                        </p>
                    </div>

                    <div className="text-center p-6 bg-white rounded-lg shadow-md">
                        <Shield className="h-12 w-12 text-purple-600 mx-auto mb-4" />
                        <h3 className="text-xl font-semibold mb-3">Secure Payments</h3>
                        <p className="text-gray-600">
                            Two-phase booking system with secure payment processing
                        </p>
                    </div>
                </div>
            </section>

            {/* Trending Events */}
            <section className="py-16">
                <div className="flex justify-between items-center mb-8">
                    <div>
                        <h2 className="text-3xl font-bold text-gray-900 mb-2">
                            Trending Events
                        </h2>
                        <p className="text-gray-600">
                            Discover what's popular right now
                        </p>
                    </div>
                    <Link
                        to="/events"
                        className="text-blue-600 hover:text-blue-700 font-semibold inline-flex items-center"
                    >
                        View All <ArrowRight className="h-4 w-4 ml-1" />
                    </Link>
                </div>

                {loading ? (
                    <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
                        {[...Array(6)].map((_, i) => (
                            <div key={i} className="bg-white rounded-lg shadow-md p-6 animate-pulse">
                                <div className="h-4 bg-gray-200 rounded mb-4"></div>
                                <div className="h-3 bg-gray-200 rounded mb-2"></div>
                                <div className="h-3 bg-gray-200 rounded mb-4"></div>
                                <div className="h-8 bg-gray-200 rounded"></div>
                            </div>
                        ))}
                    </div>
                ) : trendingEvents.length > 0 ? (
                    <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
                        {trendingEvents.map((event) => (
                            <div key={event.event_id} className="bg-white rounded-lg shadow-md overflow-hidden hover:shadow-lg transition-shadow">
                                <div className="p-6">
                                    <div className="flex items-start justify-between mb-3">
                                        <h3 className="text-lg font-semibold text-gray-900 line-clamp-2">
                                            {event.name}
                                        </h3>
                                        <Star className="h-5 w-5 text-yellow-500 flex-shrink-0 ml-2" />
                                    </div>

                                    <p className="text-gray-600 text-sm mb-2">
                                        {event.venue_name}
                                    </p>

                                    <p className="text-gray-500 text-sm mb-4">
                                        {event.venue_city}
                                    </p>

                                    <div className="flex justify-between items-center">
                                        <span className="text-lg font-bold text-blue-600">
                                            ${event.base_price}
                                        </span>
                                        <span className="text-sm text-gray-500">
                                            {event.available_seats} seats left
                                        </span>
                                    </div>

                                    <Link
                                        to={`/events/${event.event_id}`}
                                        className="mt-4 w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 transition-colors inline-flex items-center justify-center"
                                    >
                                        View Details
                                    </Link>
                                </div>
                            </div>
                        ))}
                    </div>
                ) : (
                    <div className="text-center py-12">
                        <Calendar className="h-16 w-16 text-gray-400 mx-auto mb-4" />
                        <p className="text-gray-500 text-lg">
                            No trending events available at the moment
                        </p>
                        <Link
                            to="/events"
                            className="mt-4 inline-flex items-center text-blue-600 hover:text-blue-700"
                        >
                            Browse all events <ArrowRight className="h-4 w-4 ml-1" />
                        </Link>
                    </div>
                )}
            </section>

            {/* Call to Action */}
            <section className="py-16 bg-gray-100 rounded-lg">
                <div className="text-center max-w-3xl mx-auto px-6">
                    <h2 className="text-3xl font-bold text-gray-900 mb-4">
                        Ready to Get Started?
                    </h2>
                    <p className="text-gray-600 text-lg mb-8">
                        Join thousands of users who trust BookMyEvent for their event booking needs
                    </p>

                    <div className="flex flex-col sm:flex-row gap-4 justify-center">
                        {!isAuthenticated && !isAdminAuthenticated ? (
                            <>
                                <Link
                                    to="/register"
                                    className="bg-blue-600 text-white px-8 py-3 rounded-lg font-semibold hover:bg-blue-700 transition-colors inline-flex items-center justify-center"
                                >
                                    <Users className="h-5 w-5 mr-2" />
                                    Sign Up Now
                                </Link>
                                <Link
                                    to="/admin/register"
                                    className="border-2 border-purple-600 text-purple-600 px-8 py-3 rounded-lg font-semibold hover:bg-purple-600 hover:text-white transition-colors inline-flex items-center justify-center"
                                >
                                    <Shield className="h-5 w-5 mr-2" />
                                    Admin Registration
                                </Link>
                            </>
                        ) : (
                            <Link
                                to="/events"
                                className="bg-blue-600 text-white px-8 py-3 rounded-lg font-semibold hover:bg-blue-700 transition-colors inline-flex items-center justify-center"
                            >
                                <Search className="h-5 w-5 mr-2" />
                                Discover Events
                            </Link>
                        )}
                    </div>
                </div>
            </section>
        </div>
    );
};

export default Home;
