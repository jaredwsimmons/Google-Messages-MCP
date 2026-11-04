const GA_MEASUREMENT_ID = process.env.NEXT_PUBLIC_GA_MEASUREMENT_ID?.trim() || "";
const GA_API_SECRET = process.env.GA4_API_SECRET?.trim() || "";

function parseGaClientId(cookieValue) {
  if (!cookieValue) {
    return null;
  }

  const parts = cookieValue.split(".");

  if (parts.length < 4) {
    return null;
  }

  return `${parts.at(-2)}.${parts.at(-1)}`;
}

function createFallbackClientId() {
  return `${Date.now()}.${Math.floor(Math.random() * 1_000_000_000)}`;
}

function normalizeParamValue(value) {
  if (value == null) {
    return undefined;
  }

  const stringValue = `${value}`.trim();
  return stringValue ? stringValue.slice(0, 100) : undefined;
}

export function getGoogleAnalyticsMeasurementId() {
  return GA_MEASUREMENT_ID;
}

export async function trackGoogleAnalyticsEvent({ request, name, params = {} }) {
  if (!GA_MEASUREMENT_ID || !GA_API_SECRET) {
    return false;
  }

  const clientId =
    parseGaClientId(request.cookies.get("_ga")?.value) || createFallbackClientId();
  const eventParams = Object.fromEntries(
    Object.entries(params)
      .map(([key, value]) => [key, normalizeParamValue(value)])
      .filter(([, value]) => value !== undefined)
  );
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 1200);

  try {
    const response = await fetch(
      `https://www.google-analytics.com/mp/collect?measurement_id=${encodeURIComponent(GA_MEASUREMENT_ID)}&api_secret=${encodeURIComponent(GA_API_SECRET)}`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify({
          client_id: clientId,
          events: [
            {
              name,
              params: {
                ...eventParams,
                engagement_time_msec: 1
              }
            }
          ]
        }),
        signal: controller.signal
      }
    );

    if (!response.ok) {
      console.error("Google Analytics event failed:", response.status, response.statusText);
    }

    return response.ok;
  } catch (error) {
    if (error?.name !== "AbortError") {
      console.error("Google Analytics event failed:", error);
    }

    return false;
  } finally {
    clearTimeout(timeout);
  }
}
