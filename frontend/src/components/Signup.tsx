import { useState, ChangeEvent, FormEvent, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';
import { toastifyMessage, MessageType, handleErrorWihtToast } from './Toastify.tsx';
import logo_with_text from '/logo-with-text.png';
import '../styles/Auth.css';

function Signup() {
    const [username, setUsername] = useState<string>('');
    const [email, setEmail] = useState<string>('');
    const [password, setPassword] = useState<string>('');
    const [repeatPassword, setRepeatPassword] = useState<string>('');
    const [showPassword, setShowPassword] = useState<boolean>(false);
    const [showPasswordAgain, setShowPasswordAgain] = useState<boolean>(false);
    const messageRef = useRef<HTMLDivElement>(null);
    const navigate = useNavigate();

    const handleShowPasswordChange = () => {
        setShowPassword(!showPassword);
    };

    const handleShowPasswordAgainChange = () => {
        setShowPasswordAgain(!showPasswordAgain);
    };

    const validateInputs = () => {
        const usernameRegex = /^[A-Za-z0-9]{4,12}$/;
        const emailRegex = /^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9-]+(?:.[a-zA-Z0-9-]+)*$/;
        const passwordRegex = /^(?=.*[a-z])(?=.*[A-Z])(?=.*[0-9])(?=.*[@#$!%*?&])[A-Za-z0-9@#$!%*?&]{8,32}$/;

        if (!usernameRegex.test(username)) {
            toastifyMessage("Invalid username format.", MessageType.Error);
            return false;
        }
        if (!emailRegex.test(email)) {
            toastifyMessage("Invalid email format.", MessageType.Error);
            return false;
        }
        if (!passwordRegex.test(password)) {
            toastifyMessage("Invalid password format.", MessageType.Error);
            return false;
        }

        return true;
    };

    const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
        e.preventDefault();

        if (password != repeatPassword) {
            toastifyMessage("Passwords do not match.", MessageType.Error);
            return;
        }

        if (!validateInputs()) {
            return;
        }

        try {
            const response = await axios.post('/api/signup', {
                username: username,
                email: email,
                password: password
            });

            toastifyMessage("Registered succussfully.", MessageType.Success);

            console.log('Response:', response.data);

            navigate('/login');

        } catch (error: any) {
            handleErrorWihtToast(error, "Failed to register.")
        }
    };

    const goToLogin = () => {
        navigate('/login');
    };

    return (
        <form className="auth" onSubmit={handleSubmit}>
            <div className="logo-container">
                <img src={logo_with_text} alt="logo" />
            </div>
            <label className="title-lable">Sign up to continue</label>
            <div ref={messageRef}></div>
            <div className="formRow">
                <div>
                    <label htmlFor="username">Username</label>
                    <input type="text" id="username" value={username} onChange={(e: ChangeEvent<HTMLInputElement>) => setUsername(e.target.value)} />
                </div>
                <div>
                    <label htmlFor="email">Email</label>
                    <input type="text" id="email" value={email} onChange={(e: ChangeEvent<HTMLInputElement>) => setEmail(e.target.value)} />
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
                <div>
                    <label htmlFor="repeatPassword">Repeat Password</label>
                    <input type={showPasswordAgain ? "text" : "password"} id="repeatPassword" value={repeatPassword} onChange={(e: ChangeEvent<HTMLInputElement>) => setRepeatPassword(e.target.value)} />
                </div>
                <div className="checkbox-container">
                    <input
                        type="checkbox"
                        id="showpassAgain"
                        checked={showPasswordAgain}
                        onChange={handleShowPasswordAgainChange}
                    />
                    <label htmlFor="showpassAgain">Show password</label>
                </div>
            </div>
            <button type="submit" className="submit-button">Sign Up</button>
            <div className="link" onClick={goToLogin}>
                Already have an account? Log in
            </div>
        </form>
    );
}

export default Signup;
