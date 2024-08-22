import { useState, ChangeEvent, FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';
import { toastifyMessage, MessageType, handleErrorWihtToast } from './Toastify.tsx';
import logo_with_text from '/logo-with-text.png';
import '../styles/Auth.css';

function Login() {
    const [username, setUsername] = useState<string>('');
    const [password, setPassword] = useState<string>('');
    const [showPassword, setShowPassword] = useState<boolean>(false);
    const navigate = useNavigate();

    const handleShowPasswordChange = () => {
        setShowPassword(!showPassword);
    };

    const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
        e.preventDefault();

        try {
            const response = await axios.post('/api/login', {
                username: username,
                password: password
            });

            toastifyMessage("Logged in successfully.", MessageType.Success);

            const accessToken: string = response.data.AccessToken;
            const refreshToken: string = response.data.RefreshToken;

            localStorage.setItem('accessToken', accessToken);
            localStorage.setItem('refreshToken', refreshToken);

            axios.defaults.headers.common['Authorization'] = `Bearer ${accessToken}`;

            navigate('/dashboard');
        } catch (error: any) {
            handleErrorWihtToast(error, "Failed to log in.")
        }
    };

    const goToSignup = () => {
        navigate('/signup');
    };

    return (
        <form className="auth" onSubmit={handleSubmit}>
            <div className="logo-container">
                <img src={logo_with_text} alt="logo" />
            </div>
            <label className="title-lable">Log in to continue</label>
            <div className="formRow">
                <div>
                    <label htmlFor="username">Username</label>
                    <input type="text" id="username" value={username} onChange={(e: ChangeEvent<HTMLInputElement>) => setUsername(e.target.value)} />
                </div>
                <div>
                    <label htmlFor="password">Password</label>
                    <input type={showPassword ? "text" : "password"} id="password" value={password} onChange={(e: ChangeEvent<HTMLInputElement>) => setPassword(e.target.value)} />
                </div>
                <div className="checkbox-container">
                    <input
                        type="checkbox"
                        id="showpass"
                        checked={showPassword}
                        onChange={handleShowPasswordChange}
                    />
                    <label htmlFor="showpass">Show password</label>
                </div>
            </div>
            <button type="submit" className="submit-button">Login</button>
            <div className="link" onClick={goToSignup}>
                Don't have an account? Sign Up
            </div>
        </form>
    );
}

export default Login;
