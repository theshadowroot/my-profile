(() => {
    const HEARTBEAT_INTERVAL = 15000;

    let visitId = null;
    let startTime = null;
    let heartbeatTimer = null;

    async function startVisit() {
        try {
            const response = await fetch("/api/visits", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
            });

            if (!response.ok) {
                console.error("Analytics: failed to start visit");
                return;
            }

            const data = await response.json();

            visitId = data.id;
            startTime = Date.now();

            heartbeatTimer = setInterval(updateDuration, HEARTBEAT_INTERVAL);
        } catch (error) {
            console.error("Analytics: failed to start visit", error);
        }
    }

    async function updateDuration() {
        if (!visitId || !startTime) {
            return;
        }

        const durationSeconds = Math.floor(
            (Date.now() - startTime) / 1000
        );

        try {
            await fetch(`/api/visits/${visitId}`, {
                method: "PATCH",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({
                    duration_seconds: durationSeconds,
                }),
            });
        } catch (error) {
            console.error(
                "Analytics: failed to update duration",
                error
            );
        }
    }

    function finalUpdate() {
        if (!visitId || !startTime) {
            return;
        }

        const durationSeconds = Math.floor(
            (Date.now() - startTime) / 1000
        );

        fetch(`/api/visits/${visitId}`, {
            method: "PATCH",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                duration_seconds: durationSeconds,
            }),
            keepalive: true,
        });

        if (heartbeatTimer) {
            clearInterval(heartbeatTimer);
            heartbeatTimer = null;
        }
    }

    window.addEventListener("pagehide", finalUpdate);

    startVisit();
})();
