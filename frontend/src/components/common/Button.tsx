import { Component, JSX } from 'solid-js';
import { Dynamic } from 'solid-js/web';

interface ButtonProps extends JSX.ButtonHTMLAttributes<HTMLButtonElement> {
    variant?: 'primary' | 'secondary' | 'danger' | 'ghost';
    size?: 'small' | 'medium' | 'large';
    isLoading?: boolean;
    as?: keyof JSX.IntrinsicElements;
}

const Button: Component<ButtonProps> = (props) => {
    const {
        variant = 'primary',
        size = 'medium',
        isLoading,
        children,
        class: className,
        as = 'button',
        ...rest
    } = props;

    const baseClasses = 'button';
    const variantClasses = {
        primary: 'is-primary',
        secondary: 'is-info',
        danger: 'is-danger',
        ghost: 'is-ghost',
    };
    const sizeClasses = {
        small: 'is-small',
        medium: '',
        large: 'is-large',
    };

    const classes = [
        baseClasses,
        variantClasses[variant],
        sizeClasses[size],
        isLoading ? 'is-loading' : '',
        className,
    ].join(' ');

    return (
        <Dynamic
            component={as}
            class={classes}
            disabled={isLoading}
            {...rest}
        >
            {children}
        </Dynamic>
    );
};

export default Button;