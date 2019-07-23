# golang-ucp-sim

Only supports the following operations:
* Alert - always returns success
* Submit Short Message - only returns failure when max window size is reached. Current implementation doesn't support delivery receipt yet.
* Session Management - always returns success

# Environment

* UCP_SIM_SILENT - disables logging
* UCP_SIM_CONF_DIR - config directory

# How to build 
1. Clone or get a copy of this repository
2. Run in project root
```
go build 
```

Note: implemented using Go Modules.

# Command line arguments
 ```
-o	File where outgoing ucp requests(currently msisdn is logged) are written
-r	File where receiving ucp requests(currently msisdn is logged) are written 
```


Note: implemented using Go Modules.

# HTTP Api

## Deliver Short Message
* http://localhost:8090/api/messages/deliverBulk

```
{
   "Requests":[  
      {  
         "AccessCode":"acode",  
         "Recipient":"recipient1",  
         "Message":"message1"  
      },  
      {  
         "AccessCode":"acode",  
         "Recipient":"recipient2",  
         "Message":"message2"  
      }  
   ]
}
```

## TPS

* http://localhost:8090/api/setMaxTPS
* http://localhost:8090/api/incomingTPS
* http://localhost:8090/api/successTPS
* http://localhost:8090/api/failTPS
