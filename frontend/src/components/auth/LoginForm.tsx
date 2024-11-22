import { Component, createSignal } from 'solid-js';
import { useNavigate } from '@solidjs/router';
import Input from '../common/Input';
import Button from '../common/Button';
import { authStore } from '@/stores/auth.ts';

const LoginForm: Component = () => {
    const navigate = useNavigate();
    const [email, setEmail] = createSignal('');
    const [password, setPassword] = createSignal('');

    const handleSubmit = async (e: Event) => {
        e.preventDefault();
        await authStore.login(email(), password());
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
                type="password"
                label="Password"
                value={password()}
                onInput={(e) => setPassword(e.currentTarget.value)}
                required
            />
            <Button
                type="submit"
                class="is-fullwidth"
                isLoading={authStore.state.isLoading}
            >
                Log In
            </Button>
            {authStore.state.error && (
                <p class="help is-danger mt-2">{authStore.state.error}</p>
            )}
        </form>
    );
};

export default LoginForm;