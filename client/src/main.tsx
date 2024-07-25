import React from "react";
import ReactDOM from "react-dom/client";
import "./index.css";
import "react-toastify/dist/ReactToastify.css";
import { Slide, ToastContainer } from "react-toastify";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import { GroupsOverview } from "./views/GroupsOverview.tsx";
import { GroupView } from "./views/GroupView.tsx";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

const queryClient = new QueryClient();

const App = () => {
	return (
		<div className="mx-auto lg:w-5/6 w-[95%] pt-10">
			<Routes>
				<Route path="/" element={<GroupsOverview />} />
				<Route
					path="/group/archived/:groupId"
					element={<GroupView archived />}
				/>
				<Route path="/group/:groupId" element={<GroupView />} />
				<Route path="*" element={<div>Not Found</div>} />
			</Routes>
		</div>
	);
};

// biome-ignore lint/style/noNonNullAssertion: <explanation>
ReactDOM.createRoot(document.getElementById("root")!).render(
	<React.StrictMode>
		<BrowserRouter>
			<QueryClientProvider client={queryClient}>
				<App />
			</QueryClientProvider>

			<ToastContainer
				position="top-right"
				autoClose={3000}
				hideProgressBar={false}
				newestOnTop
				closeOnClick
				rtl={false}
				pauseOnFocusLoss
				pauseOnHover
				theme="dark"
				transition={Slide}
			/>
		</BrowserRouter>
	</React.StrictMode>,
);
