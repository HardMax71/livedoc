import { Component } from 'solid-js';
import LoginForm from '../../components/auth/LoginForm';
import { A } from '@solidjs/router';

const Login: Component = () => {
    return (
        <div class="container">
            <div class="columns is-centered">
                <div class="column is-5-tablet is-4-desktop">
                    <div class="box mt-6">
                        <h1 class="title has-text-centered">Log In</h1>
                        <LoginForm />
                        <p class="has-text-centered mt-4">
                            Don't have an account? <A href="/register">Register</A>
                        </p>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Login;