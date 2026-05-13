import { useState } from 'react';
import { AlertTriangle, X } from 'lucide-react';

export default function SecurityBanner() {
    const isInsecure = window.location.protocol !== 'https:';
    const [dismissed, setDismissed] = useState(() => {
        try {
            return window.localStorage.getItem('security-banner-dismissed') === 'true';
        } catch {
            return false;
        }
    });

    const handleDismiss = () => {
        setDismissed(true);
        try {
            window.localStorage.setItem('security-banner-dismissed', 'true');
        } catch {
            // ignore storage write errors
        }
    };

    if (!isInsecure || dismissed) {
        return null;
    }

    return (
        <>
            {/* Spacer to prevent content overlap */}
            <div className="h-12 sm:h-10" />

            <div className="fixed top-0 left-0 right-0 z-[9999] bg-gradient-to-r from-red-600 to-orange-600 text-white py-2 px-3 sm:px-4 shadow-lg">
                <div className="container mx-auto flex items-center justify-between gap-2 sm:gap-4">
                    <div className="flex items-center gap-2 sm:gap-3 min-w-0">
                        <AlertTriangle className="w-4 h-4 sm:w-5 sm:h-5 flex-shrink-0" />
                        <span className="text-xs sm:text-sm font-medium truncate sm:whitespace-normal">
                            <span className="font-bold">Not Secure:</span>
                            <span className="hidden sm:inline"> This connection is not encrypted (HTTP). Your sensitive data is not fully protected.</span>
                            <span className="sm:hidden"> HTTP connection!</span>
                        </span>
                    </div>
                    <button
                        onClick={handleDismiss}
                        className="text-white/80 hover:text-white transition-colors p-1 flex-shrink-0"
                        aria-label="Dismiss security warning"
                    >
                        <X className="w-4 h-4" />
                    </button>
                </div>
            </div>
        </>
    );
}
