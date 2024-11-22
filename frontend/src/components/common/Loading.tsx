import { Component } from 'solid-js';

interface LoadingProps {
    size?: 'small' | 'medium' | 'large';
}

const Loading: Component<LoadingProps> = (props) => {
    const { size = 'medium' } = props;

    const sizeClasses = {
        small: 'is-small',
        medium: '',
        large: 'is-large',
    };

    return (
        <div class="has-text-centered p-4">
            <span class={`loader ${sizeClasses[size]}`}></span>
        </div>
    );
};

export default Loading;