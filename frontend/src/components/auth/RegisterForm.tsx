import { Component, createSignal } from 'solid-js';
import { useNavigate } from '@solidjs/router';
import Input from '../common/Input';
import Button from '../common/Button';
import { authStore } from '@/stores/auth.ts';

const RegisterForm: Component = () => {
    const navigate = useNavigate();
    const [email, setEmail] = createSignal('');
    const [username, setUsername] = createSignal('');
    const [password, setPassword] = createSignal('');
    const [confirmPassword, setConfirmPassword] = createSignal('');
    const [passwordError, setPasswordError] = createSignal('');

    const handleSubmit = async (e: Event) => {
        e.preventDefault();

        if (password() !== confirmPassword()) {
            setPasswordError('Passwords do not match');
            return;
        }

        await authStore.register(email(), username(), password());
        if (authStore.state.isAuthenticated) {
            navigate('/documents');
        }
    };

    return (
        <form onSubmit={handleSubmit} class="auth-form">
            <Input
                type="email"
                label="Email"
                value={email()}
                onInput={(e) => setEmail(e.currentTarget.value)}
                required
            />
            <Input
                type="text"
                label="Username"
                value={username()}
                onInput={(e) => setUsername(e.currentTarget.value)}
                required
            />
            <Input
                type="password"
                label="Password"
                value={password()}
                onInput={(e) => setPassword(e.currentTarget.value)}
                required
            />
            <Input
                type="password"
                label="Confirm Password"
                value={confirmPassword()}
                onInput={(e) => setConfirmPassword(e.currentTarget.value)}
                error={passwordError()}
                required
            />
            <Button
                type="submit"
                class="is-fullwidth"
                isLoading={authStore.state.isLoading}
            >
                Register
            </Button>
            {authStore.state.error && (
                <p class="help is-danger mt-2">{authStore.state.error}</p>
            )}
        </form>
    );
};

export default RegisterForm;