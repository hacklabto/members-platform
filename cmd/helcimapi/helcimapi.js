module.exports.findCustomer = async ({name, email}) => {
	console.log("looking for customer", name, email)

	let res = await fetch("https://api.helcim.com/v2/customers/", {
		headers: { "api-token": process.env.NEW_TOKEN }
	});

	let customers = await res.json();
	// console.log(customers);

	let customerByEmail = !!email ? customers.find(c => c.billingAddress?.email === email || c.shippingAddress?.email === email) : null;
	let customerByName = !!name ? customers.find(c => c.contactName === name || c.billingAddress?.name === name || c.shippingAddress?.name === name) : null;

	if (!customerByEmail && !customerByName) {
		console.log("could not find customer!");
		return null;
	}

	if (!customerByEmail) return customerByName;
	if (!customerByName) return customerByEmail;

	if (customerByEmail.customerCode != customerByName.customerCode) {
		throw { userVisibleError: "Matched different customers by name and email!" };
	}

	return customerByEmail;
}

const { XMLParser } = require("fast-xml-parser");

module.exports.findSubscription = async (customerID) => {
	console.log("querying subscriptions for customer " + customerID);

	let res = await fetch("https://secure.myhelcim.com/api/recurring/subscription-search", {
		method: "POST",
		body: `customerCode=${customerID}`,
		headers: {
			"account-id": process.env.ACCOUNT_ID,
			"api-token": "thisdoesntmatter",
			"content-type": "application/x-www-form-urlencoded",
		}
	});
	
	let xml = await res.text();

	const parsert = new XMLParser();
	let unxmled = parsert.parse(xml)?.subscriptions?.subscription;

	console.log(unxmled);

	console.log(typeof unxmled)

	if (Array.isArray(unxmled)) {
		return unxmled.find(s => s.recurringPlanCode === "member1st" && s.status === "Active");
	}

	if (typeof unxmled === 'object') {
		if (unxmled.recurringPlanCode === "member1st" && unxmled.status === "Active") {
			return unxmled;
		}
		return null;
	}

	if (!unxmled) return null;

	console.log(unxmled);
	throw 'unknown unxmled type';
}