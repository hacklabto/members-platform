require("dotenv").config();

const http = require("http");

const { findCustomer, findSubscription } = require("./helcimapi.js");

const getBody = (r) => new Promise(res => {
	let data = '';
	r.on('data', (chunk) => data += chunk);
	r.on('end', () => res(data));
})

const process = async (data) => {
	if (!data.name && !data.email) {
		throw { userVisibleError: "must have either 'name' or 'email' in request" };
	}

	let customer = await findCustomer(data);
	if (!customer) return [200, { ok: false }]; //throw { userVisibleError: "Customer record not found" };
	console.log("found customer", customer);

	let subscription = await findSubscription(customer.customerCode);
	console.log("found subscription?", subscription);
	return [200, { ok: !!subscription }]
}

const handler = async (req, res) => {
	console.log(req.method, req.url, req.headers);
	let data = await getBody(req);
	
	if (!data) return;
	try {
		data = JSON.parse(data);
	} catch (e) {
		console.error(e);
		res.statusCode = 400;
		res.end();
		return;
	}

	console.log("d", data);

	let status, body;

	try {
		[status, body] = await process(data);
	} catch(e) {
		if (e?.userVisibleError) {
			status = 400;
			body = { error: e.userVisibleError };
		} else {
			console.error(e);
			status = 500;
			body = {"error": "500 internal server error"}
		}
	}

	res.statusCode = status;
	res.write(JSON.stringify(body));
	res.end();
};

http.createServer(handler).listen(8080, "0.0.0.0")