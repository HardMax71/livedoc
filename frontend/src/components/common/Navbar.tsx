import { Component } from 'solid-js';
import { A } from '@solidjs/router';
import { authStore } from '../../stores/auth';

const Navbar: Component = () => {
    const handleLogout = async () => {
        await authStore.logout();
    };

    return (
        <nav class="navbar is-primary" role="navigation" aria-label="main navigation">
            <div class="navbar-brand">
                <A class="navbar-item" href="/">
                    <strong>SyncWrite</strong>
                </A>

                <a
                    role="button"
                    class="navbar-burger"
                    aria-label="menu"
                    aria-expanded="false"
                    data-target="navbarBasic"
                    href="#"
                >
                    <span aria-hidden="true"></span>
                    <span aria-hidden="true"></span>
                    <span aria-hidden="true"></span>
                </a>
            </div>

            <div id="navbarBasic" class="navbar-menu">
                <div class="navbar-start">
                    {authStore.state.isAuthenticated && (
                        <A class="navbar-item" href="/documents">
                            My Documents
                        </A>
                    )}
                </div>

                <div class="navbar-end">
                    <div class="navbar-item">
                        <div class="buttons">
                            {!authStore.state.isAuthenticated ? (
                                <>
                                    <A class="button is-light" href="/register">
                                        Sign up
                                    </A>
                                    <A class="button is-light" href="/login">
                                        Log in
                                    </A>
                                </>
                            ) : (
                                <button class="button is-light" onClick={handleLogout}>
                                    Log out
                                </button>
                            )}
                        </div>
                    </div>
                </div>
            </div>
        </nav>
    );
};

export default Navbar;
