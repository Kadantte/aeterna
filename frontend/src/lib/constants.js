export const ALLOWED_EXTENSIONS = ['.pdf', '.txt', '.doc', '.docx', '.jpg', '.jpeg', '.png', '.gif', '.webp', '.zip'];
export const MAX_FILE_SIZE = 10 * 1024 * 1024; // 10 MB
export const MAX_FILES = 5;
export const MAX_TOTAL_SIZE = 25 * 1024 * 1024; // 25 MB
export const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export const TIME_PRESETS = [
    { label: '1 Minute (Debug)', value: 1 },
    { label: '15 Minutes (Test)', value: 15 },
    { label: '1 Hour', value: 60 },
    { label: '1 Day', value: 1440 },
    { label: '3 Days', value: 4320 },
    { label: '1 Week', value: 10080 },
    { label: '2 Weeks', value: 20160 },
    { label: '1 Month', value: 43200 },
    { label: '3 Months', value: 129600 },
    { label: '6 Months', value: 259200 },
    { label: '1 Year', value: 525600 },
];

export const REMINDER_PRESETS = [
    { label: '15 Minutes Before', value: 15 },
    { label: '1 Hour Before', value: 60 },
    { label: '12 Hours Before', value: 720 },
    { label: '1 Day Before', value: 1440 },
    { label: '2 Days Before', value: 2880 },
    { label: '3 Days Before', value: 4320 },
    { label: '5 Days Before', value: 7200 },
    { label: '10 Days Before', value: 14400 },
];

export const FAREWELL_DELAY_PRESETS = [
    { label: 'Immediately after trigger', value: 0 },
    { label: '1 hour after trigger', value: 60 },
    { label: '6 hours after trigger', value: 360 },
    { label: '12 hours after trigger', value: 720 },
    { label: '1 day after trigger', value: 1440 },
    { label: '3 days after trigger', value: 4320 },
    { label: '1 week after trigger', value: 10080 },
    { label: '2 weeks after trigger', value: 20160 },
];