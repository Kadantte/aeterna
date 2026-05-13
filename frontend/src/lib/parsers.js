export function parseRecipientEmails(input) {
    const parts = String(input || '')
        .split(/[\n,;]+/)
        .map((entry) => entry.trim())
        .filter(Boolean);

    const seen = new Set();
    const unique = [];
    for (const email of parts) {
        const key = email.toLowerCase();
        if (!seen.has(key)) {
            seen.add(key);
            unique.push(email);
        }
    }
    return unique;
}