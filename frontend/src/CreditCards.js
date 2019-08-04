import React from 'react';
import { injectStripe, StripeProvider, Elements, CardElement } from 'react-stripe-elements';

const INITIALSTATE = "INITIAL", SUCCESSSTATE = "COMPLETE", FAILEDSTATE = "FAILED";

class CreditCardForm extends React.Component{
    constructor(props){
        super(props);

        this.state = {
            status : INITIALSTATE,
            value : '',
        };

        this.handleChange = this.handleChange.bind(this);
        this.handleSubmit = this.handleSubmit.bind(this);
    }
    
    handleChange(event) {
        event.preventDefault();
        const name = event.target.name;
        const value = event.target.value;
        console.log("handleChange :" + event.target.name)
        this.setState({
            [name]: value
        });
    }

    async handleSubmit(event) {
        event.preventDefault();
        let id = "";

        if (this.props.stripe) {
            console.log('Strip is loaded');
        } else {
            console.log('Form submitted before Stripe loaded.');
        }
       //If we are not using a pre-saved card, connect with stripe to obtain a card token
        if (!this.state.useExisting) {
           //Create the token via Stripe's API
           console.log(this.state.name);
            let { token } = await this.props.stripe.createToken({ name: this.state.name});
            if (token == null) {
                console.log("invalid token");
                this.setState({ status: FAILEDSTATE });
                return;
            }
            id = token.id;
            console.log("Stripe Token ID" + id);
        }
        //Create the request, then send it to the back-end
        let response = await fetch("/users/charge", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
                token: id,
                customer_id: this.props.user,
                product_id: this.props.productid,
                sell_price: this.props.price,
                rememberCard: this.state.remember !== undefined,
                useExisting: this.state.useExisting
            })
        });
        //If response is ok, consider the operation a success
        if (response.ok) {
            console.log("Purchase Complete!");
            this.setState({ status: SUCCESSSTATE });
        } else {
            this.setState({ status: FAILEDSTATE });
        }
    }

    renderCreditCardInformation() {
        const style = {
            base: {
                'fontSize': '20px',
                'color': '#495057',
                'fontFamily': 'apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,"Helvetica Neue",Arial,sans-serif'
            }
        };
        const usersavedcard = <div>
            <div className="form-row text-center">
                <button type="button" className="btn btn-outline-success text-center mx-auto">Use saved card?</button>
            </div>
            <hr />
        </div>

        const remembercardcheck = <div className="form-row form-check text-center">
            <input className="form-check-input" type="checkbox" value="" id="remembercardcheck" onChange={this.handleChange} />
            <label className="form-check-label" htmlFor="remembercardcheck">
                Remember Card?
            </label>
        </div>;
        return (
            <div>
                {usersavedcard}
                <h5 className="mb-4">Payment Info</h5>
                <form onSubmit={this.handleSubmit}>
                    <div className="form-row">
                        <div className="col-lg-12 form-group">
                            <label htmlFor="cc-name">Name On Card:</label>
                            <input id="cc-name" name='name' className="form-control" placeholder='Name on Card' onChange={this.handleChange} type='text' />
                        </div>
                    </div>
                    <div className="form-row">
                        <div className="col-lg-12 form-group">
                            <label htmlFor="card">Card Information:</label>
                            <CardElement id="card" className="form-control" style={style} />
                        </div>
                    </div>
                    {remembercardcheck}
                    <hr className="mb-4" />
                    <button type="submit" className="btn btn-success btn-large" >{this.props.operation}</button>
                </form>
            </div>
        );  
    }
    renderSuccess() {
        return (
            <div>
                <h5 className="mb-4 text-success">Request Successfull....</h5>
                <button type="submit" className="btn btn-success btn-large" onClick={() => { this.props.toggle() }}>Ok</button>
            </div>
        );
    }
    renderFailure(){
        return (
            <div>
                <h5 className="mb-4 text-danger"> Credit card information invalid, try again or exit</h5>
                {this.renderCreditCardInformation()}
            </div>
        );
    } 

    render() {
        let body = null;

        switch (this.state.status) {
            case SUCCESSSTATE:
                body = this.renderSuccess();
                break;
            case FAILEDSTATE:
                body = this.renderFailure();
                break;
            default:
                body = this.renderCreditCardInformation();
        }

        return (
            <div>
                {body}
            </div>
        );
    }
}

export default function CreditCardInformation(props) {
    if (!props.show) {
        return null;
    }

    console.log("Inside CreditCardInformation")

    const CCFormWithStripe = injectStripe(CreditCardForm);
    return (
        <div>
            {props.separator ? <hr /> : null}
            <StripeProvider apiKey="pk_test_TYooMQauvdEDq54NiTphI7jx">
                <Elements>
                    <CCFormWithStripe user={props.user} operation={props.operation} productid={props.productid} price={props.price} toggle={props.toggle} />
                </Elements>
            </StripeProvider>
        </div>
    );
}
