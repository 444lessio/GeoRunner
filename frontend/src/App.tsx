// Import necessary libraries from React
import React, { useState, useEffect } from 'react';
// Import components from the 'react-leaflet' library
import { MapContainer, TileLayer, Marker, Popup } from 'react-leaflet';
// Import the type for map coordinates from 'leaflet'
import { LatLngExpression } from 'leaflet';

// --- Leaflet Icon Fix ---
// This entire block is a standard workaround for a known issue
// where 'create-react-app' and Webpack can't find the default marker icons.
import L from 'leaflet';
import icon from 'leaflet/dist/images/marker-icon.png';
import iconShadow from 'leaflet/dist/images/marker-shadow.png';

// Create a new default icon instance with the correct image paths
let DefaultIcon = L.icon({
    iconUrl: icon,
    shadowUrl: iconShadow,
    iconSize: [25, 41],   // Adjust the size
    iconAnchor: [12, 41] // Adjust the anchor point
});

// Globally override Leaflet's default marker icon
L.Marker.prototype.options.icon = DefaultIcon;
// --- End of Icon Fix ---


// Define a TypeScript interface for the structure of our Driver data
// This ensures type safety when we fetch data from the API
interface Driver {
  id: string;
  lat: number;
  lon: number;
}

// This is the main React component
function App() {
  // Initialize the component's state using the 'useState' hook
  // 'drivers' will hold the array of driver objects
  // 'setDrivers' is the function we use to update this array
  const [drivers, setDrivers] = useState<Driver[]>([]);
  
  // Define a constant for the map's initial center position (0,0 is the equator)
  const mapCenter: LatLngExpression = [0, 0]; 

  // Define an asynchronous function to fetch driver data from our Go backend
  const fetchDrivers = async () => {
    try {
      // Call our API endpoint.
      // Note: This is hardcoded to always ask for drivers near (lat=0, lon=0)
      // 'searchRadius' is set on the backend in 'main.go'
      const response = await fetch("http://localhost:8080/find-nearby?lat=0&lon=0");
      
      // Basic error handling if the network request fails
      if (!response.ok) {
        throw new Error('Network response was not ok');
      }
      
      // Parse the JSON response into our 'Driver[]' type
      const data: Driver[] = await response.json();
      
      // Update the component's state with the new driver data
      // This will cause React to re-render the map with the new pins
      setDrivers(data); 
    } catch (error) {
      // Log any errors to the browser console for debugging
      console.error("Error fetching drivers:", error);
    }
  };

  // Use the 'useEffect' hook to run code when the component first loads
  // The empty array '[]' at the end means "run this effect only once on mount"
  useEffect(() => {
    // 1. Fetch the driver data immediately when the component loads
    fetchDrivers(); 
    
    // 2. Set up an interval to automatically call 'fetchDrivers' every 2000ms (2 seconds)
    // This creates the "real-time" updating effect
    const intervalId = setInterval(fetchDrivers, 2000); 

    // 3. This 'return' function is a "cleanup" function.
    // React runs this when the component is unmounted (e.g., user leaves the page).
    // This is crucial to prevent memory leaks by stopping the interval.
    return () => clearInterval(intervalId);
  }, []); // The empty '[]' dependency array ensures this effect runs only once

  // This is the JSX (HTML-like) code that React will render
  return (
    <MapContainer 
      center={mapCenter} 
      zoom={3} // Set the initial zoom level to see a large part of the world
      style={{ height: '100vh', width: '100%' }} // Make the map fill the entire screen
    >
      {/* This component renders the actual map "tiles" (the images)
          We get them from OpenStreetMap */}
      <TileLayer
        url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
      />

      {/* This is the dynamic part:
          We loop (.map) over the 'drivers' array from our state. */}
      {drivers.map(driver => (
        // For each driver, we render a <Marker> (pin) component
        // 'key' is a special React prop for lists, helping it track items
        <Marker key={driver.id} position={[driver.lat, driver.lon]}>
          {/* Add a popup that appears when you click the pin */}
          <Popup>
            Driver ID: {driver.id}
          </Popup>
        </Marker>
      ))}
    </MapContainer>
  );
}

// Export the 'App' component so 'index.tsx' can import and render it
export default App;