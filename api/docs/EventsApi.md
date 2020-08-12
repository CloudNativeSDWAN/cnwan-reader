# \EventsApi

All URIs are relative to *http://localhost/cnwan*

Method | HTTP request | Description
------------- | ------------- | -------------
[**SendEvents**](EventsApi.md#SendEvents) | **Post** /events | Last observed events



## SendEvents

> Response SendEvents(ctx, event)

Last observed events

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**event** | [**[]Event**](Event.md)| List of observed events | 

### Return type

[**Response**](Response.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

