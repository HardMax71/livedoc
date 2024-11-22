import { Component } from 'solid-js';
import { useNavigate } from '@solidjs/router';
import Button from '../components/common/Button';
import { authStore } from '../stores/auth';

const Home: Component = () => {
    const navigate = useNavigate();

    return (
        <div class="hero is-fullheight-with-navbar">
            <div class="hero-body">
                <div class="container has-text-centered">
                    <h1 class="title is-1">
                        SyncWrite
                    </h1>
                    <h2 class="subtitle is-3">
                        Real-time collaborative document editing
                    </h2>
                    <div class="buttons is-centered mt-6">
                        {!authStore.state.isAuthenticated ? (
                            <>
                                <Button
                                    size="large"
                                    onClick={() => navigate('/register')}
                                >
                                    Get Started
                                </Button>
                                <Button
                                    variant="ghost"
                                    size="large"
                                    onClick={() => navigate('/login')}
                                >
                                    Log In
                                </Button>
                            </>
                        ) : (
                            <Button
                                size="large"
                                onClick={() => navigate('/documents')}
                            >
                                My Documents
                            </Button>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Home;