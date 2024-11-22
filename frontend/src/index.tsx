import { render } from 'solid-js/web';
import { Router } from '@solidjs/router';
import App from './App';

import '@fontsource/inter/400.css';
import '@fontsource/inter/500.css';
import '@fontsource/inter/600.css';
import '@fontsource/inter/700.css';
import './styles/main.scss';
import './styles/editor.scss';

const root = document.getElementById('root');

if (import.meta.env.DEV && !(root instanceof HTMLElement)) {
    throw new Error(
        'Root element not found. Did you forget to add it to your index.html?',
    );
}

render(
    () => (
        <Router>
            <App />
        </Router>
    ),
    root!,
);