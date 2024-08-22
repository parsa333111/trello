import { useEffect, useState } from 'react';
import { BrowserRouter as Router, Routes, Route, useNavigate, useLocation } from 'react-router-dom';
import axios from 'axios';
import Signup from './components/Signup';
import Login from './components/Login';
import Dashboard from './components/Dashboard';
import Workspace from './components/Workspace';

function App() {
    const location = useLocation();
    const navigate = useNavigate();
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        const rootElement = document.getElementById('root');
        if (rootElement) {
            if (location.pathname.startsWith('/workspace/')) {
                rootElement.className = 'workspace-route';
            } else {
                switch (location.pathname) {
                    case '/signup':
                        rootElement.className = 'signup-route';
                        break;
                    case '/login':
                        rootElement.className = 'login-route';
                        break;
                    case '/dashboard':
                        rootElement.className = 'dashboard-route';
                        break;
                    default:
                        navigate('/login');
                        return;
                }
            }
        }

        const authenticate = async () => {
            const accessToken = localStorage.getItem('accessToken');
            const refreshToken = localStorage.getItem('refreshToken');

            try {
                if (accessToken) {
                    axios.defaults.headers.common['Authorization'] = `Bearer ${accessToken}`;

                    await axios.get('/api/token/validate');

                    console.log("Access token validated.");

                    if (location.pathname === '/login' || location.pathname === '/signup') {
                        navigate('/dashboard');
                    }

                } else if (refreshToken) {
                    throw new Error('Access token is missing or invalid');

                } else {
                    if (location.pathname !== '/login' && location.pathname !== '/signup') {
                        navigate('/login');
                    }
                }
            } catch (error) {
                console.error('Access token validation failed:', error);

                localStorage.removeItem('accessToken');

                try {
                    if (refreshToken) {
                        axios.defaults.headers.common['RefreshToken'] = refreshToken;

                        const response = await axios.post('/api/refresh-token', { token: refreshToken });
                        const newAccessToken = response.data.accessToken;

                        localStorage.setItem('accessToken', newAccessToken);
                        axios.defaults.headers.common['Authorization'] = `Bearer ${newAccessToken}`;

                        console.log("Token refreshed.");

                        if (location.pathname === '/login' || location.pathname === '/signup') {
                            navigate('/dashboard');
                        }
                    } else {
                        if (location.pathname !== '/login' && location.pathname !== '/signup') {
                            navigate('/login');
                        }
                    }
                } catch (error) {
                    console.error('Token refresh failed:', error);

                    localStorage.removeItem('refreshToken');

                    if (location.pathname !== '/login' && location.pathname !== '/signup') {
                        navigate('/login');
                    }
                }
            }
            finally {
                setIsLoading(false);
            }
        };

        authenticate();
    }, [location, navigate]);

    if (isLoading) {
        return null;
    }

    return (
        <Routes>
            <Route path="/login" element={<Login />} />
            <Route path="/signup" element={<Signup />} />
            <Route path="/dashboard" element={<Dashboard />} />
            <Route path="/workspace/:workspace_id" element={<Workspace />} />
            {/* Add more routes here */}
        </Routes>
    );
}

function RootApp() {
    return (
        <Router>
            <App />
        </Router>
    );
}

export default RootApp;
