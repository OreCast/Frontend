// https://stackoverflow.com/questions/14229695/google-maps-api-throws-uncaught-referenceerror-google-is-not-defined-only-whe
//
// Initialize and add the map
let map;

async function initMap() {
  // The location of Uluru
//  const position = { lat: -25.344, lng: 131.031 };
  // position of Ithaca NY
  const position = { lat: 42.443962, lng: -76.501884 };
  // Request needed libraries.
  //@ts-ignore
  const { Map } = await google.maps.importLibrary("maps");
  const { AdvancedMarkerElement } = await google.maps.importLibrary("marker");

  // The map, centered at Uluru
  map = new Map(document.getElementById("map"), {
    zoom: 4,
    center: position,
    mapId: "DEMO_MAP_ID",
  });

  // The marker, positioned at Uluru
  const marker = new AdvancedMarkerElement({
    map: map,
    position: position,
    title: "Cornell University",
  });
}

initMap();
