import { FAREWELL_DELAY_PRESETS } from './constants.js';

export function formatFileSize(bytes) {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
}

export function formatMinutes(minutes) {
    if (minutes >= 1440) {
        const days = Number((minutes / 1440).toFixed(1));
        return `${days} Day${days !== 1 ? 's' : ''} Before`;
    }
    if (minutes >= 60) {
        const hours = Number((minutes / 60).toFixed(1));
        return `${hours} Hour${hours !== 1 ? 's' : ''} Before`;
    }
    return `${minutes} Minutes Before`;
}

export function formatFarewellDelay(minutes) {
    const preset = FAREWELL_DELAY_PRESETS.find(p => p.value === minutes);
    if (preset) return preset.label;
    if (minutes === 0) return 'Immediately after trigger';
    if (minutes >= 10080) return `${minutes / 10080} week(s) after trigger`;
    if (minutes >= 1440) return `${minutes / 1440} day(s) after trigger`;
    if (minutes >= 60) return `${minutes / 60} hour(s) after trigger`;
    return `${minutes} minute(s) after trigger`;
}