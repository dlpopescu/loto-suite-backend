include .env.mk

# .PHONY: upload-config
# upload-config:
# ifdef CONFIG_FILES
# 	scp -i $(EC2_KEY) $(CONFIG_FILES) $(EC2_USER)@$(EC2_IP):~
# endif

# SSH into EC2
.PHONY: ssh
ssh:
	ssh -i $(EC2_KEY) $(EC2_USER)@$(EC2_IP)

.PHONY: deploy-backend
deploy-backend: 
	$(MAKE) -f Makefile.backend deploy-backend
	@echo "Backend deployment complete."

.PHONY: deploy
deploy:
	$(MAKE) -f Makefile.cleanup cleanup-services
	$(MAKE) -f Makefile.cleanup cleanup-binaries
	$(MAKE) -f Makefile.backend deploy-backend
	$(MAKE) -f Makefile.redirect deploy-redirect
	@echo "Full deployment complete."