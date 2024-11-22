import { Component } from 'solid-js';
import { Router, Route } from '@solidjs/router';
import Navbar from './components/common/Navbar';
import Home from './pages/Home';
import Login from './pages/auth/Login';
import Register from './pages/auth/Register';
import DocumentList from './pages/documents/List';
import EditDocument from './pages/documents/Edit';
import NewDocument from './pages/documents/New';
import NotFound from './pages/NotFound';
import { authStore } from './stores/auth';

interface ProtectedRouteProps {
    component: Component;
}

const ProtectedRoute: Component<ProtectedRouteProps> = (props) => {
    if (!authStore.state.isAuthenticated) {
        return <Login />;
    }
    const ProtectedComponent = props.component;
    return <ProtectedComponent />;
};

const App: Component = () => {
    return (
        <Router>
            <Navbar />
            <Route path="/" component={Home} />
            <Route path="/login" component={Login} />
            <Route path="/register" component={Register} />

            <Route
                path="/documents"
                component={() => (
                    <ProtectedRoute component={DocumentList} />
                )}
            />
            <Route
                path="/documents/new"
                component={() => (
                    <ProtectedRoute component={NewDocument} />
                )}
            />
            <Route
                path="/documents/:id"
                component={() => (
                    <ProtectedRoute component={EditDocument} />
                )}
            />

            {/* Catch-all Route for 404 Not Found */}
            <Route path="*" component={NotFound} />
        </Router>
    );
};

export default App;
