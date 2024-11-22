import { Component } from 'solid-js';
import RegisterForm from '../../components/auth/RegisterForm';
import { A } from '@solidjs/router';

const Register: Component = () => {
    return (
        <div class="container">
            <div class="columns is-centered">
                <div class="column is-5-tablet is-4-desktop">
                    <div class="box mt-6">
                        <h1 class="title has-text-centered">Create Account</h1>
                        <RegisterForm />
                        <p class="has-text-centered mt-4">
                            Already have an account? <A href="/login">Log in</A>
                        </p>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Register;