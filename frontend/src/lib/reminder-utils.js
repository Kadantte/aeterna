export function applyDurationToReminders(reminders, newDuration) {
    const validReminders = reminders.filter((value) => value < newDuration);
    if (validReminders.length > 0) {
        return validReminders;
    }

    if (newDuration >= 1440) {
        return [newDuration / 2];
    }
    if (newDuration >= 60) {
        return [15];
    }
    return [];
}

export function addReminderValue(reminders, value, maxDuration) {
    if (reminders.includes(value) || value >= maxDuration) {
        return reminders;
    }
    return [...reminders, value].sort((a, b) => b - a);
}

export function removeReminderValue(reminders, value) {
    return reminders.filter((item) => item !== value);
}
