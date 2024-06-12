import { useEffect } from "react";
import "./App.css";
import axios from "axios";

function App() {
  useEffect(() => {
    axios.get(`/api/scrape`).then((data) => {
      console.log(data);
    });
  }, []);

  return <>hello worldasdfljadslfkj</>;
}

export default App;
