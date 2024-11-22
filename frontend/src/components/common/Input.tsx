import { Component, JSX } from 'solid-js';

interface InputProps extends JSX.InputHTMLAttributes<HTMLInputElement> {
    label?: string;
    error?: string;
    helperText?: string;
}

const Input: Component<InputProps> = (props) => {
    const {
        label,
        error,
        helperText,
        class: className,
        ...rest
    } = props;

    return (
        <div class="field">
            {label && (
                <label class="label">{label}</label>
            )}
            <div class="control">
                <input
                    class={`input ${error ? 'is-danger' : ''} ${className || ''}`}
                    {...rest}
                />
            </div>
            {(error || helperText) && (
                <p class={`help ${error ? 'is-danger' : ''}`}>
                    {error || helperText}
                </p>
            )}
        </div>
    );
};

export default Input;