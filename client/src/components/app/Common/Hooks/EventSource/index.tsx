import { KcEventSource } from "@/types";
import { isIP } from "@/utils";
import { useEffect, useMemo } from "react";

const useEventSource = ({ url, sendMessage }: KcEventSource) => {
  const updatedUrl = useMemo(() => {
    if (window.location.protocol === "http:") {
      if (!isIP(window.location.host.split(":")[0])) {
        return `http://${new Date().getTime()}.${window.location.host}${url}`;
      }
      return `http://${window.location.host}${url}`;
    }
    return url;
  }, [url]);

  useEffect(() => {
    // opening a connection to the server to begin receiving events from it
    const eventSource = new EventSource(updatedUrl);
    // attaching a handler to receive message events
    eventSource.onmessage = (event) => {
      // const eventData = JSON.parse(event.data);
      // sendMessage(eventData)

      try {
        const eventData = JSON.parse(event.data);
        sendMessage(eventData);
      } catch {
        // const eventData = JSON.parse(event.data);
        sendMessage(event.data);
      }
    };
    
    // terminating the connection on component unmount
    return () => eventSource.close();
  }, [updatedUrl, sendMessage]);
};

export {
  useEventSource
};
