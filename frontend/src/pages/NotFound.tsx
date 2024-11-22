import { Component } from 'solid-js';
import { A } from '@solidjs/router';

const NotFound: Component = () => {
    return (
        <div class="hero is-fullheight-with-navbar">
            <div class="hero-body">
                <div class="container has-text-centered">
                    <h1 class="title is-1">404</h1>
                    <h2 class="subtitle">Page not found</h2>
                    <A href="/" class="button is-primary">
                        Go Home
                    </A>
                </div>
            </div>
        </div>
    );
};

export default NotFound;