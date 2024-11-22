import { format, formatDistance } from 'date-fns';

export const formatDate = (date: Date | string): string => {
    const d = new Date(date);
    return format(d, 'MMM dd, yyyy HH:mm');
};

export const formatRelative = (date: Date | string): string => {
    const d = new Date(date);
    return formatDistance(d, new Date(), { addSuffix: true });
};

export const formatTime = (date: Date | string): string => {
    const d = new Date(date);
    return format(d, 'HH:mm:ss');
};